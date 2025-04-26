package openai

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
)

func GenerateSchema[T any]() map[string]any {
	ref := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}

	var v T
	s := ref.Reflect(v) // *jsonschema.Schema

	// marshal → unmarshal to turn it into a generic map
	raw, _ := json.Marshal(s)

	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return m
}
