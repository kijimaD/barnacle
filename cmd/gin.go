package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

var CmdGin = &cli.Command{
	Name:        "gin",
	Usage:       "Ginを起動する",
	Description: "Ginを起動する",
	Action:      initGin,
	Flags:       []cli.Flag{},
}

// FIXME: ミドルウェア入れたいだけなのに、ハンドラもセットされてるよな…
func initGin(ctx *cli.Context) error {
	r := gin.Default()
	myv := MyValidator{}
	r.Use(myv.MyMiddleware())
	r.GET("/square/:n", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"result": 100,
		})
	})
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"result": "hello",
		})
	})
	r.GET("/not_impl", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"result": "hello",
		})
	})

	r.Run()
	return nil
}

type MyValidator struct{}

func (v MyValidator) MyMiddleware() gin.HandlerFunc {
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
	if err := doc.Validate(context.Background()); err != nil { // Assert our OpenAPI is valid!
		panic(err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		route, pathParams, err := router.FindRoute(c.Request)
		if err != nil {
			c.Abort()
			return
		}

		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
			Options:    &openapi3filter.Options{},
		}
		if err = openapi3filter.ValidateRequest(c.Request.Context(), requestValidationInput); err != nil {
			fmt.Println(err)
			c.Abort()
			return
		}

		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w // 書かないと動かない

		c.Next()

		if err = openapi3filter.ValidateResponse(c.Request.Context(), &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 c.Writer.Status(),
			Header:                 c.Writer.Header(),
			Body:                   io.NopCloser(w.body),
			Options:                &openapi3filter.Options{},
		}); err != nil {
			fmt.Println(err)
			c.Writer.Write([]byte(`"{error:レスポンスがAPI仕様を満たしていない}"`))
			c.Abort()
			return
		}
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
