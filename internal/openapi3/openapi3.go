package openapi3

import (
	"strings"
)

// SchemaError -.
type SchemaError struct {
	Ref string
}

// Error -.
func (e *SchemaError) Error() string {
	return "unknown schema " + e.Ref
}

// OpenAPI Object
// See specification https://swagger.io/specification/#openapi-object
type OpenAPI struct {
	Info       Info       `json:"info" yaml:"info"`
	Paths      Paths      `json:"paths" yaml:"paths"`
	Components Components `json:"components,omitempty" yaml:"components,omitempty"`
}

// LookupByReference -.
func (api OpenAPI) LookupByReference(ref string) (Schema, error) {
	schema := api.Components.Schemas[schemaKey(ref)]
	if nil == schema {
		return Schema{}, &SchemaError{Ref: ref}
	}

	return *schema, nil
}

func schemaKey(ref string) string {
	const prefix = "#/components/schemas/"
	return strings.TrimPrefix(ref, prefix)
}
