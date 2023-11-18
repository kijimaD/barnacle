package cmd

import (
	"barnacle/barnacle"
	"bytes"
	"context"
	"io"
	"log"

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
	r.Use(MyMiddleware())
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
	// API仕様を満たしていないレスポンス
	r.GET("/not_impl", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"result": "hello",
		})
	})
	r.Run()

	return nil
}

const APIResposeErrorCode = "API RESPONSE ERROR"
const APIResposeErrorMsg = "レスポンスがAPI仕様を満たしていない"

func MyMiddleware() gin.HandlerFunc {
	doc, err := openapi3.NewLoader().LoadFromData([]byte(barnacle.Docstr[1:]))
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
			log.Println(err)
			c.Abort()
			return
		}

		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w
		c.Next()

		if err = openapi3filter.ValidateResponse(c.Request.Context(), &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 c.Writer.Status(),
			Header:                 c.Writer.Header(),
			Body:                   io.NopCloser(w.body),
			Options:                &openapi3filter.Options{},
		}); err != nil {
			log.Println(err)
			c.JSON(200, gin.H{
				"code": APIResposeErrorCode,
				"msg":  APIResposeErrorMsg,
			})
			c.String(200, "\n"+err.Error()) // ステータスコードは500にしたいが、反映されない
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
