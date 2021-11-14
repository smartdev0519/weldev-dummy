package test

import (
	"github.com/go-dummy/dummy/internal/config"
	"github.com/go-dummy/dummy/internal/openapi3"
	"github.com/go-dummy/dummy/internal/server"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheck(t *testing.T) {
	cases, err := ioutil.ReadDir("./case")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		response, err := ioutil.ReadFile("case/" + c.Name() + "/response.json")
		if err != nil {
			t.Fatal(err)
		}

		path := "case/" + c.Name() + "/openapi3.yaml"

		openapi, err := openapi3.Parse(path)
		if err != nil {
			t.Fatal(err)
		}

		s := new(server.Server)
		s.Cfg = config.Server{
			Port: "4000",
			Path: path,
		}

		s.OpenAPI = openapi

		s.Handlers = make(map[string]server.Handler)

		if err := s.SetHandlers(); err != nil {
			t.Fatal(err)
		}

		mux := http.NewServeMux()

		mux.HandleFunc("/", s.Handler)

		newServer := httptest.NewServer(mux)

		for k, v := range s.OpenAPI.Paths {
			if v.Get != nil {
				resp, err := http.Get(newServer.URL + k)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				out, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}

				require.JSONEq(t, string(out), string(response))
			}
		}

	}
}
