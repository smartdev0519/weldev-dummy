package openapi3_test

import (
	"github.com/go-dummy/dummy/internal/openapi3"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse_Yaml(t *testing.T) {
	expected := openapi3.OpenAPI{
		Info: openapi3.Info{
			Title:   "Users dummy API",
			Version: "0.1.0",
		},
		Paths: openapi3.Paths{
			"/users": &openapi3.Path{
				Post: &openapi3.Operation{
					Parameters: nil,
					Responses: openapi3.Responses{
						"201": &openapi3.Response{
							Description: ptrStr(""),
							Content: openapi3.Content{
								"application/json": &openapi3.MediaType{
									Example:  nil,
									Examples: nil,
								},
							},
						},
					},
				},
				Get: &openapi3.Operation{
					Parameters: nil,
					Responses: openapi3.Responses{
						"200": &openapi3.Response{
							Description: ptrStr(""),
							Content: openapi3.Content{
								"application/json": &openapi3.MediaType{
									Example: []any{
										map[any]any{
											"id":        "e1afccea-5168-4735-84d4-cb96f6fb5d25",
											"firstName": "Elon",
											"lastName":  "Musk",
										},
										map[any]any{
											"id":        "472063cc-4c83-11ec-81d3-0242ac130003",
											"firstName": "Sergey",
											"lastName":  "Brin",
										},
									},
									Examples: nil,
								},
							},
						},
					},
				},
			},
			"/users/{userId}": &openapi3.Path{
				Get: &openapi3.Operation{
					Parameters: openapi3.Parameters{
						openapi3.Parameter{
							Name:     "userId",
							In:       "path",
							Required: true,
							Schema: &openapi3.Schema{
								Type: "string",
							},
						},
					},
					Responses: openapi3.Responses{
						"200": &openapi3.Response{
							Description: ptrStr(""),
							Content: openapi3.Content{
								"application/json": &openapi3.MediaType{},
							},
						},
					},
				},
			},
		},
		Components: openapi3.Components{
			Schemas: openapi3.Schemas{
				"User": &openapi3.Schema{
					Properties: openapi3.Schemas{
						"id": &openapi3.Schema{
							Type:   "string",
							Format: "uuid",
						},
						"firstName": &openapi3.Schema{
							Type: "string",
						},
						"lastName": &openapi3.Schema{
							Type: "string",
						},
					},
					Type: "object",
				},
			},
		},
	}

	openapi, err := openapi3.Parse("testdata/openapi3.yml")

	require.NoError(t, err)
	require.Equalf(t, &expected, openapi, `parsed schema from "testdata/openapi3.yml"`)
}

func ptrStr(s string) *string {
	return &s
}
