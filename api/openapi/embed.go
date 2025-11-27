package openapi

import (
	_ "embed"
)

//go:embed hakoniwa.yaml
var openAPISchema []byte

func GetOpenAPIAISchema() []byte {
	return openAPISchema
}
