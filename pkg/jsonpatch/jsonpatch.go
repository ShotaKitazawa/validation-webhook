package jsonpatch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ShotaKitazawa/validation-webhook/pkg/errors"
)

const (
	apiVersion = "admission.k8s.io/v1beta1"
)

func JsonPatch(uid string, violate error) (map[string]interface{}, error) {
	var respStr string

	switch customErr := violate.(type) {
	case *errors.Immutable:
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
`, apiVersion, uid, 400, strings.ReplaceAll(customErr.Error(), `"`, `\"`))

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
`, apiVersion, uid)
	}

	var jsonBody map[string]interface{}
	if err := json.Unmarshal([]byte(respStr), &jsonBody); err != nil {
		err = fmt.Errorf("JSON Unmarshal error: %s", err)
		return nil, err
	}
	return jsonBody, nil
}
