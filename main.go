package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func main() {
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
                    type: string
                required: [result]
  /hello:
    get:
      description: hello
      responses:
        '200':
          description: hello
          content:
            "application/json":
              schema:
                type: object
                properties:
                  result:
                    type: string
                required: [result]
`[1:]))
	if err != nil {
		panic(err)
	}

	squareHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		result := map[string]interface{}{"result": "square"}
		if err = json.NewEncoder(w).Encode(&result); err != nil {
			panic(err)
		}
	})

	// Start an http server.
	var mainHandler http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 先にhttptest.ServerのURLを確定させるため
		mainHandler.ServeHTTP(w, r)
	}))
	defer srv.Close()

	// Patch the OpenAPI spec to match the httptest.Server.URL. Only needed
	// because the server URL is dynamic here.
	doc.Servers = []*openapi3.Server{{URL: srv.URL}}
	if err := doc.Validate(context.Background()); err != nil { // Assert our OpenAPI is valid!
		panic(err)
	}
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}
	v := openapi3filter.NewValidator(router, openapi3filter.Strict(true),
		openapi3filter.OnErr(func(w http.ResponseWriter, status int, code openapi3filter.ErrCode, err error) {
			// カスタムバリデーション
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
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(resp.StatusCode, strings.TrimSpace(string(contents)))
	}
	// Valid requests to our sum service
	printResp(srv.Client().Get(srv.URL + "/square/2"))
	printResp(srv.Client().Get(srv.URL + "/hello")) // 処理はsquareHandler...
	printResp(srv.Client().Get(srv.URL + "/"))      // error
}
