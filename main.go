package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type config struct {
	certFile string
	keyFile  string
}

func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")

	fl.Parse(os.Args[1:])
	return cfg
}

//func Handler(w http.ResponseWriter, r *http.Request) {
func Handler(c echo.Context) error {
	//if r.Header.Get("Content-Type") != "application/json" {
	//	w.WriteHeader(http.StatusBadRequest) // 400
	//	return
	//}

	r := c.Request()
	if r.Body == nil {
		return c.String(http.StatusBadRequest, "body is nil") // 400
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "failed to read body") // 399
	}

	ar, err := newAdmissionReview(body)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error()) // 400
	}

	violate, err := Validation(ar)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error()) // 400
	}

	resp, err := jsonPatch(ar, violate)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error()) // 500
	}

	c.JSON(http.StatusOK, resp)
	return nil
}

func main() {
	cfg := initFlags()

	e := echo.New()

	// Middleware
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		fmt.Printf("{\"time\": \"%s\", ", time.Now().String())
		fmt.Printf("\"request_body\": %v, ", strings.ReplaceAll(string(reqBody), "\n", ""))
		fmt.Printf("\"response_body\": %v", strings.ReplaceAll(string(resBody), "\n", ""))
		fmt.Printf("}\n")
	}))

	// Routes
	e.POST("/", Handler)

	if err := http.ListenAndServeTLS(":8080", cfg.certFile, cfg.keyFile, e); err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
