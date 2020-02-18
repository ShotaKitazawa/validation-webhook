package main

import (
	"encoding/json"
	"fmt"
	"strings"

	admission "k8s.io/api/admission/v1beta1"
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
	var object, oldObject map[string]interface{}
	if err := json.Unmarshal(ar.Request.Object.Raw, &object); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(ar.Request.OldObject.Raw, &oldObject); err != nil {
		return nil, err
	}

	path := `ingress.metadata.annotations["kubernetes.io/ingress.global-static-ip-name"]`
	escapedPath, err := escape(path)
	if err != nil {
		return nil, nil
	}
	provideGIP, err := search(object, escapedPath)
	if err != nil {
		return nil, nil
	}
	currentGIP, err := search(oldObject, escapedPath)
	if err != nil {
		return nil, nil
	}

	if currentGIP != "" && provideGIP != currentGIP {
		return &Immutable{Field: `Ingress.metadata.annotations["kubernetes.io/ingress.global-static-ip-name"]`}, nil
	}

	return nil, nil
}

func escape(str string) (string, error) {
	head := strings.Index(str, "[")
	tail := strings.Index(str, "]")

	// search end
	if head < 0 && tail < 0 {
		return str, nil
	}
	// invalid
	if head < 0 || tail < 0 {
		return "", fmt.Errorf("invalid syntax")
	}
	a := str[:head]
	c, err := escape(str[tail+1:])
	if err != nil {
		return "", err
	}
	b := strings.ReplaceAll(strings.ReplaceAll(str[head+1:tail], ".", "&pe"), "\"", "")
	if c == "" {
		return a + "." + b, nil
	} else {
		return a + "." + b + "." + c, nil
	}

}
func unescape(str string) string {
	return strings.ReplaceAll(str, "&pe", ".")
}

func search(obj map[string]interface{}, path string) (interface{}, error) {
	topField := strings.Split(path, ".")[0]
	if strings.ToLower(obj["kind"].(string)) != strings.ToLower(topField) {
		return nil, fmt.Errorf("no much kind")
	}
	return recursiveSearch(obj, path[strings.Index(path, ".")+1:])
}

func recursiveSearch(obj map[string]interface{}, path string) (interface{}, error) {
	topField := strings.Split(path, ".")[0]
	if path != topField {
		newObj, ok := obj[unescape(topField)]
		if !ok {
			return nil, fmt.Errorf("no much field")
		}
		newObjMAP, ok := newObj.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("this field is not map")
		}
		return recursiveSearch(newObjMAP, path[strings.Index(path, ".")+1:])
	}
	if _, ok := obj[unescape(topField)]; !ok {
		return nil, fmt.Errorf("no much field")
	}
	return obj[unescape(topField)], nil
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
`, apiVersion, ar.Request.UID, 400, strings.ReplaceAll(customErr.Error(), `"`, `\"`))

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
