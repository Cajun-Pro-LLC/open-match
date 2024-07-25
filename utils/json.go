package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

func MarshalUnsafe(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		b = []byte(err.Error())
	}
	return string(b)
}

func PrettyPrint(title string, content interface{}) {
	header := fmt.Sprintf("~~~~~%s~~~~~", title)
	var data string
	switch t := content.(type) {
	case string:
		data = t
	case []byte:
		data = string(t)
	case int, int32, int64:
		data = fmt.Sprintf("%d", t)
	case float32, float64:
		data = fmt.Sprintf("%f", t)
	default:
		data = MarshalUnsafe(content)
	}
	footer := strings.Repeat("~", len(header))
	fmt.Printf("%s\n%s\n%s\n", header, data, footer)
}
