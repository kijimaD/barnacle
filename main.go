package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func initHandler() http.Handler {
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
  /not_impl:
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
                  type:
                    type: string
                required: [result, type]
`[1:]))
	if err != nil {
		panic(err)
	}

	squareHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		result := map[string]interface{}{"result": 100}
		if err = json.NewEncoder(w).Encode(&result); err != nil {
			panic(err)
		}
	})
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		result := map[string]interface{}{"result": "hello"}
		if err = json.NewEncoder(w).Encode(&result); err != nil {
			panic(err)
		}
	})

	// Start an http server.
	var mainHandler http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 先にhttptest.ServerのURLを確定させるため
		// 外部から処理を注入できるようにする。変数mainHandlerに値を入れると、srvの動きが変わる
		mainHandler.ServeHTTP(w, r)
	}))
	defer srv.Close()

	// Patch the OpenAPI spec to match the httptest.Server.URL. Only needed
	// because the server URL is dynamic here.
	doc.Servers = []*openapi3.Server{{URL: srv.URL}, {URL: "localhost:8080"}}
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

	mux := http.NewServeMux()
	mux.HandleFunc("/square/", squareHandler)
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/not_found", helloHandler)
	mux.HandleFunc("/not_impl", helloHandler)
	// Now we can finally set the main server handler.
	mainHandler = v.Middleware(mux)
	return mainHandler
}

func main() {
	http.ListenAndServe(":8080", initHandler())
}
