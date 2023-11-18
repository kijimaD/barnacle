package main

import (
	"context"
	"encoding/json"
	"net/http"

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
                  type:
                    type: string
                required: [type]
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

	if err := doc.Validate(context.Background()); err != nil { // Assert our OpenAPI is valid!
		panic(err)
	}
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}
	v := openapi3filter.NewValidator(router, openapi3filter.Strict(true),
		openapi3filter.OnErr(func(w http.ResponseWriter, status int, code openapi3filter.ErrCode, err error) {
			// カスタムレスポンス
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

	var mainHandler http.Handler
	mainHandler = v.Middleware(mux)
	return mainHandler
}

func main() {
	http.ListenAndServe(":8080", initHandler())
}
