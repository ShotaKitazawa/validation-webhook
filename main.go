package main

import (
	"io/ioutil"
	"net/http"

	admission "k8s.io/api/admission/v1"
	//appsv1 "k8s.io/api/apps/v1"
)

func newAdmissionReview(body []byte) (*admission.AdmissionReview, error) {
	ar := admission.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		return nil, err
	}
	return &ar, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) // 405
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	var body []byte
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	ar, err := newAdmissionReview(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	violate, err := Validation(ar)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	resp, err := jsonPatch(ar, violate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func main() {
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":8080", nil)
}
