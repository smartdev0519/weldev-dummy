package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-dummy/dummy/internal/openapi3"
)

type Handler struct {
	Path       string
	Method     string
	QueryParam url.Values
	Header     map[string]string
	StatusCode int
	Response   interface{}
}

func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Set-Status-Code") == "500" {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if h, ok := s.GetHandler(r.Method, RemoveTrailingSlash(r.URL.Path), r.URL.Query(), r.Header.Get("X-Example"), r.Body); ok {
		w.WriteHeader(h.StatusCode)
		bytes, _ := json.Marshal(h.Response)
		_, _ = w.Write(bytes)

		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) SetHandlers() error {
	for path, method := range s.OpenAPI.Paths {
		if method.Get != nil {
			handlers, err := handlers(path, http.MethodGet, method.Get)
			if err != nil {
				return err
			}

			s.Handlers[path] = append(s.Handlers[path], handlers...)
		}

		if method.Post != nil {
			handlers, err := handlers(path, http.MethodPost, method.Post)
			if err != nil {
				return err
			}

			s.Handlers[path] = append(s.Handlers[path], handlers...)
		}

		if method.Put != nil {
			handlers, err := handlers(path, http.MethodPut, method.Put)
			if err != nil {
				return err
			}

			s.Handlers[path] = append(s.Handlers[path], handlers...)
		}

		if method.Patch != nil {
			handlers, err := handlers(path, http.MethodPatch, method.Patch)
			if err != nil {
				return err
			}

			s.Handlers[path] = append(s.Handlers[path], handlers...)
		}
	}

	return nil
}

func handlers(path, method string, o *openapi3.Operation) ([]Handler, error) {
	var res []Handler

	queryParam := make(url.Values)

	for i := 0; i < len(o.Parameters); i++ {
		if o.Parameters[i].In == "query" {
			queryParam.Add(o.Parameters[i].Name, "")
		}
	}

	for code, resp := range o.Responses {
		statusCode, err := strconv.Atoi(code)
		if err != nil {
			return nil, err
		}

		content := resp.Content["application/json"]

		examplesKeys := content.Examples.GetExamplesKeys()

		if len(examplesKeys) > 0 {
			res = append(res, handler(path, method, queryParam, map[string]string{}, statusCode, content.ResponseByExamplesKey(examplesKeys[0])))

			for i := 0; i < len(examplesKeys); i++ {
				res = append(res, handler(path, method, queryParam, map[string]string{"example": examplesKeys[i]}, statusCode, content.ResponseByExamplesKey(examplesKeys[i])))
			}
		} else {
			res = append(res, handler(path, method, queryParam, map[string]string{}, statusCode, content.ResponseByExample()))
		}
	}

	return res, nil
}

func handler(path, method string, queryParam url.Values, header map[string]string, statusCode int, response interface{}) Handler {
	return Handler{
		Path:       path,
		Method:     method,
		QueryParam: queryParam,
		Header:     header,
		StatusCode: statusCode,
		Response:   response,
	}
}

func (s *Server) GetHandler(method, path string, queryParam url.Values, exampleHeader string, body io.ReadCloser) (Handler, bool) {
	for mask, handlers := range s.Handlers {
		if PathByParamDetect(path, mask) {
			for i := 0; i < len(handlers); i++ {
				if handlers[i].Method == method {
					header, ok := handlers[i].Header["example"]
					if ok && header == exampleHeader {
						return handlers[i], true
					}
				}
			}

			for i := 0; i < len(handlers); i++ {
				if handlers[i].Method == method {
					if LastPathSegmentIsParam(mask) && handlers[i].Response == nil {
						for _, v := range s.Handlers[ParentPath(mask)] {
							if v.Method == method {
								data := v.Response.([]map[string]interface{})
								for i := 0; i < len(data); i++ {
									if data[i]["id"] == GetLastPathSegment(path) {
										s.Handlers[path] = append(s.Handlers[path], handler(path, method, url.Values{}, map[string]string{}, 200, data[i]))

										return s.Handlers[path][0], true
									}
								}

								return Handler{}, false
							}
						}
					}

					if method == http.MethodPost {
						for i := 0; i < len(s.Handlers[path]); i++ {
							if s.Handlers[path][i].Method == http.MethodGet {
								data, ok := s.Handlers[path][i].Response.([]map[string]interface{})
								if ok {
									var res map[string]interface{}

									err := json.NewDecoder(body).Decode(&res)
									if err != nil {
										s.Logger.Log().Err(err)
									}

									data = append(data, res)

									s.Handlers[path][i].Response = data

									return Handler{
										StatusCode: http.StatusCreated,
										Response:   res,
									}, true
								}
							}
						}
					}

					limit, offset, found, err := pagination(queryParam)
					if err != nil {
						s.Logger.Log().Err(err)
					}

					if found {
						data := handlers[i].Response.([]map[string]interface{})
						if offset > len(data) {
							return Handler{}, true
						}

						size := len(data) - offset
						if size > limit {
							size = limit
						}

						resp := make([]map[string]interface{}, size)
						n := 0

						for i := offset; n < size; i++ {
							resp[n] = data[i]
							n++
						}

						return Handler{
							Response:   resp,
							StatusCode: http.StatusOK,
						}, true
					}

					return handlers[i], true
				}
			}
		}
	}

	return Handler{}, false
}

func PathByParamDetect(path, param string) bool {
	p := strings.Split(path, "/")
	m := strings.Split(param, "/")

	if len(p) != len(m) {
		return false
	}

	for i := 0; i < len(p); i++ {
		if strings.HasPrefix(m[i], "{") && strings.HasSuffix(m[i], "}") {
			continue
		}

		if p[i] != m[i] {
			return false
		}
	}

	return true
}

func ParentPath(path string) string {
	p := strings.Split(path, "/")

	return strings.Join(p[0:len(p)-1], "/")
}

func LastPathSegmentIsParam(path string) bool {
	p := strings.Split(path, "/")

	return strings.HasPrefix(p[len(p)-1], "{") && strings.HasSuffix(p[len(p)-1], "}")
}

func GetLastPathSegment(path string) string {
	p := strings.Split(path, "/")

	return p[len(p)-1]
}

func pagination(queryParam url.Values) (limit, offset int, found bool, err error) {
	l, ok := queryParam["limit"]
	if !ok {
		return
	}

	if ok {
		limit, err = strconv.Atoi(l[0])
	}

	o, ok := queryParam["offset"]
	if ok {
		offset, err = strconv.Atoi(o[0])
	}

	found = true

	return
}

func RemoveTrailingSlash(path string) string {
	if len(path) > 0 && path[len(path)-1] == '/' {
		return path[0 : len(path)-1]
	}

	return path
}

// RemoveFragment - clear url from reference in path
func RemoveFragment(path string) string {
	return strings.Split(path, "#")[0]
}
