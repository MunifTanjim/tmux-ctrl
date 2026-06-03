package util

import "encoding/json"

func MustJSONMarshal(data any) []byte {
	blob, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return blob
}

func MustJSONMarshalIndent(data any, prefix, indent string) []byte {
	blob, err := json.MarshalIndent(data, prefix, indent)
	if err != nil {
		panic(err)
	}
	return blob
}
