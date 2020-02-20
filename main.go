package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/ShotaKitazawa/validation-webhook/pkg/errors"
	"github.com/ShotaKitazawa/validation-webhook/pkg/jsonpatch"
	"github.com/ShotaKitazawa/validation-webhook/pkg/readfile"
	"github.com/ShotaKitazawa/validation-webhook/pkg/search"
)

var (
	cfg *config

	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func newAdmissionReview(body []byte) (*admission.AdmissionReview, error) {
	ar := admission.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		return nil, err
	}
	return &ar, nil
}

func validation(ar *admission.AdmissionReview) (violate error, err error) {
	// Immutable Check
	var object, oldObject map[string]interface{}
	if err := json.Unmarshal(ar.Request.Object.Raw, &object); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(ar.Request.OldObject.Raw, &oldObject); err != nil {
		return nil, err
	}

	lines, err := readfile.Read(cfg.policyFile)

	// TODO: evaluate multiple `path`
	// TODO: concurrency evaluate
	path := lines[0]

	escapedPath, err := search.Escape(path)
	if err != nil {
		return nil, nil
	}
	provide, err := search.Search(object, escapedPath)
	if err != nil {
		return nil, nil
	}
	current, err := search.Search(oldObject, escapedPath)
	if err != nil {
		return nil, nil
	}

	if current != "" && provide != current {
		return &errors.Immutable{Field: path}, nil
	}

	return nil, nil
}

func getUID(ar *admission.AdmissionReview) (uid string, err error) {
	uid = string(ar.Request.UID)
	return
}

func Handler(c echo.Context) error {
	r := c.Request()
	if r.Body == nil {
		return c.String(http.StatusBadRequest, "body is nil") // 400
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "failed to read body") // 400
	}

	ar, err := newAdmissionReview(body)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error()) // 400
	}

	violate, err := validation(ar)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error()) // 400
	}

	resp, err := jsonpatch.JsonPatch(string(ar.Request.UID), violate)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error()) // 500
	}

	c.JSON(http.StatusOK, resp)
	return nil
}

type config struct {
	certFile   string
	keyFile    string
	policyFile string
}

func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")
	fl.StringVar(&cfg.policyFile, "policy-file", "", "Policy file")
	fl.Parse(os.Args[1:])

	if cfg.certFile == "" || cfg.keyFile == "" || cfg.policyFile == "" {
		fmt.Println("not enough required fields")
		os.Exit(2)
	}

	return cfg
}

func main() {

	// Initialize
	cfg = initFlags()
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

	// Listen
	if err := http.ListenAndServeTLS(":8080", cfg.certFile, cfg.keyFile, e); err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
