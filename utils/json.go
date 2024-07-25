package utils

import (
	"encoding/json"
)

func MarshalUnsafe(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		b = []byte(err.Error())
	}
	return string(b)
}
