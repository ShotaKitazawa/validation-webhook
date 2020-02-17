package main

import (
	"encoding/json"
	"fmt"

	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	apiVersion = "admission.k8s.io/v1"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

type Immutable struct {
	Field string
}

func (e *Immutable) Error() string {
	return fmt.Sprintf("%s: this field is immutable", e.Field)
}

func Validation(ar *admission.AdmissionReview) (violate error, err error) {
	if violate = validationImmutable(ar); violate != nil {
		return violate, nil
	}
	return nil, nil
}

func validationImmutable(ar *admission.AdmissionReview) error {
	return &Immutable{}
}

func jsonPatch(ar *admission.AdmissionReview, violate error) ([]byte, error) {
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
	return []byte(respStr), nil
}
