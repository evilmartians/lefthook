package config

import (
	_ "embed"
)

//go:embed jsonschema.json
var JsonSchema []byte
