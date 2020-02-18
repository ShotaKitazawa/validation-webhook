package main

import (
	"encoding/json"
	"fmt"

	admission "k8s.io/api/admission/v1beta1"
	ingress "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	apiVersion = "admission.k8s.io/v1beta1"
)

var (
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

type Immutable struct {
	Field string
}

func (e *Immutable) Error() string {
	return fmt.Sprintf("%s: this field is immutable", e.Field)
}

func Validation(ar *admission.AdmissionReview) (violate error, err error) {
	// Immutable Check
	var object, oldObject ingress.Ingress
	if err := json.Unmarshal(ar.Request.Object.Raw, &object); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(ar.Request.OldObject.Raw, &oldObject); err != nil {
		fmt.Println("hoge")
		return nil, err
	}
	provideGIP := object.Annotations["kubernetes.io/ingress.global-static-ip-name"]
	currentGIP := oldObject.Annotations["kubernetes.io/ingress.global-static-ip-name"]
	if currentGIP != "" && provideGIP != currentGIP {
		return &Immutable{Field: "Ingress.metadata.annotations['kubernetes.io/ingress.global-static-ip-name']"}, nil
	}

	return nil, nil
}

func jsonPatch(ar *admission.AdmissionReview, violate error) (map[string]interface{}, error) {
	var respStr string

	switch customErr := violate.(type) {
	case *Immutable:
		respStr = fmt.Sprintf(`
{
  "apiVersion": "%[1]s",
  "kind": "AdmissionReview",
  "response": {
    "uid": "%[2]s",
    "allowed": false,
    "status": {
      "code": %[3]d,
      "message": "%[4]s"
    }
  }
}
`, apiVersion, ar.Request.UID, 400, customErr.Error())

	default:
		respStr = fmt.Sprintf(`
{
  "apiVersion": "%[1]s",
  "kind": "AdmissionReview",
  "response": {
    "uid": "%[2]s",
    "allowed": true
  }
}
`, apiVersion, ar.Request.UID)
	}

	var jsonBody map[string]interface{}
	if err := json.Unmarshal([]byte(respStr), &jsonBody); err != nil {
		err = fmt.Errorf("JSON Unmarshal error: %s", err)
		return nil, err
	}
	return jsonBody, nil
}
