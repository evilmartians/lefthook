package config

import (
	"encoding/json"

	"github.com/tidwall/jsonc"
)

type JSONC struct{}

func jsoncParser() *JSONC {
	return &JSONC{}
}

func (p *JSONC) Unmarshal(b []byte) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(jsonc.ToJSON(b), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Marshal marshals the given config map to JSON bytes.
func (p *JSONC) Marshal(o map[string]any) ([]byte, error) {
	return json.Marshal(o)
}
