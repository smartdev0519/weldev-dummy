package test

import (
	"github.com/go-dummy/dummy/internal/config"
	"github.com/go-dummy/dummy/internal/logger"
	"github.com/go-dummy/dummy/internal/openapi3"
	"github.com/go-dummy/dummy/internal/server"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	cases, err := ioutil.ReadDir("./cases")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		response, err := ioutil.ReadFile("cases/" + c.Name() + "/response.json")
		if err != nil {
			t.Fatal(err)
		}

		path := "cases/" + c.Name() + "/openapi3.yaml"

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
		s.Logger = logger.NewLogger()
		s.Handlers = make(map[string]server.Handler)

		if err := s.SetHandlers(); err != nil {
			t.Fatal(err)
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", s.Handler)
		newServer := httptest.NewServer(mux)

		for k, v := range s.OpenAPI.Paths {
			t.Run(c.Name(), func(t *testing.T) {
				var method strings.Builder
				switch {
				case v.Post != nil:
					method.WriteString(http.MethodPost)
				case v.Get != nil:
					method.WriteString(http.MethodGet)
				}

				req, err := http.NewRequest(method.String(), newServer.URL+k, nil)
				if err != nil {
					t.Fatal(err)
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}

				out, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}

				if resp.StatusCode != http.StatusNotFound {
					require.JSONEq(t, string(out), string(response))
				}
			})
		}
	}
}
