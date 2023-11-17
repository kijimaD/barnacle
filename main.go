package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func main() {
	// OpenAPI specification for a simple service that squares integers, with
	// some limitations.
	doc, err := openapi3.NewLoader().LoadFromData([]byte(`
openapi: 3.0.0
info:
  title: 'Validator - square example'
  version: '0.0.0'
paths:
  /square/{x}:
    get:
      description: square an integer
      parameters:
        - name: x
          in: path
          schema:
            type: integer
          required: true
      responses:
        '200':
          description: squared integer response
          content:
            "application/json":
              schema:
                type: object
                properties:
                  result:
                    type: integer
                    minimum: 0
                    maximum: 1000000
                required: [result]
                additionalProperties: false`[1:]))
	if err != nil {
		panic(err)
	}

	// Square service handler sanity checks inputs, but just crashes on invalid
	// requests.
	squareHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xParam := path.Base(r.URL.Path)
		x, err := strconv.ParseInt(xParam, 10, 64)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		result := map[string]interface{}{"result": x * x}
		if x == 42 {
			// An easter egg. Unfortunately, the spec does not allow additional properties...
			result["comment"] = "the answer to the ulitimate question of life, the universe, and everything"
		}
		if err = json.NewEncoder(w).Encode(&result); err != nil {
			panic(err)
		}
	})

	// Start an http server.
	var mainHandler http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Why are we wrapping the main server handler with a closure here?
		// Validation matches request Host: to server URLs in the spec. With an
		// httptest.Server, the URL is dynamic and we have to create it first!
		// In a real configured service, this is less likely to be needed.
		mainHandler.ServeHTTP(w, r)
	}))
	defer srv.Close()

	// Patch the OpenAPI spec to match the httptest.Server.URL. Only needed
	// because the server URL is dynamic here.
	doc.Servers = []*openapi3.Server{{URL: srv.URL}}
	if err := doc.Validate(context.Background()); err != nil { // Assert our OpenAPI is valid!
		panic(err)
	}
	// This router is used by the validator to match requests with the OpenAPI
	// spec. It does not place restrictions on how the wrapped handler routes
	// requests; use of gorilla/mux is just a validator implementation detail.
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}
	// Strict validation will respond HTTP 500 if the service tries to emit a
	// response that does not conform to the OpenAPI spec. Very useful for
	// testing a service against its spec in development and CI. In production,
	// availability may be more important than strictness.
	v := openapi3filter.NewValidator(router, openapi3filter.Strict(true),
		openapi3filter.OnErr(func(w http.ResponseWriter, status int, code openapi3filter.ErrCode, err error) {
			// Customize validation error responses to use JSON
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  status,
				"message": http.StatusText(status),
			})
		}))
	// Now we can finally set the main server handler.
	mainHandler = v.Middleware(squareHandler)

	printResp := func(resp *http.Response, err error) {
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(resp.StatusCode, strings.TrimSpace(string(contents)))
	}
	// Valid requests to our sum service
	printResp(srv.Client().Get(srv.URL + "/square/2"))
	printResp(srv.Client().Get(srv.URL + "/square/789"))
	// 404 Not found requests - method or path not found
	printResp(srv.Client().Post(srv.URL+"/square/2", "application/json", bytes.NewBufferString(`{"result": 5}`)))
	printResp(srv.Client().Get(srv.URL + "/sum/2"))
	printResp(srv.Client().Get(srv.URL + "/square/circle/4")) // Handler would process this; validation rejects it
	printResp(srv.Client().Get(srv.URL + "/square"))
	// 400 Bad requests - note they never reach the wrapped square handler (which would panic)
	printResp(srv.Client().Get(srv.URL + "/square/five"))
	// 500 Invalid responses
	printResp(srv.Client().Get(srv.URL + "/square/42"))    // Our "easter egg" added a property which is not allowed
	printResp(srv.Client().Get(srv.URL + "/square/65536")) // Answer overflows the maximum allowed value (1000000)
}
