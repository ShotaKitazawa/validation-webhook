package search

import (
	"fmt"
	"strings"
)

func Escape(str string) (string, error) {
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
	c, err := Escape(str[tail+1:])
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
func Unescape(str string) string {
	return strings.ReplaceAll(str, "&pe", ".")
}

func Search(obj map[string]interface{}, path string) (interface{}, error) {
	topField := strings.Split(path, ".")[0]
	if strings.ToLower(obj["kind"].(string)) != strings.ToLower(topField) {
		return nil, fmt.Errorf("no much kind")
	}
	return RecursiveSearch(obj, path[strings.Index(path, ".")+1:])
}

func RecursiveSearch(obj map[string]interface{}, path string) (interface{}, error) {
	topField := strings.Split(path, ".")[0]
	if path != topField {
		newObj, ok := obj[Unescape(topField)]
		if !ok {
			return nil, fmt.Errorf("no much field")
		}
		newObjMAP, ok := newObj.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("this field is not map")
		}
		return RecursiveSearch(newObjMAP, path[strings.Index(path, ".")+1:])
	}
	if _, ok := obj[Unescape(topField)]; !ok {
		return nil, fmt.Errorf("no much field")
	}
	return obj[Unescape(topField)], nil
}
