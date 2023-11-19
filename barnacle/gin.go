package barnacle

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
)

const APIRequestErrorCode = "API_REQUEST_ERROR"
const APIRequestErrorMsg = "リクエストがAPI仕様を満たしていない"
const APIResposeErrorCode = "API_RESPONSE_ERROR"
const APIResposeErrorMsg = "レスポンスがAPI仕様を満たしていない"

func MakeValidateMiddleware() (gin.HandlerFunc, error) {
	doc, err := openapi3.NewLoader().LoadFromData([]byte(Docstr[1:]))
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(context.Background()); err != nil { // Assert our OpenAPI is valid!
		return nil, err
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, err
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
			c.JSON(500, gin.H{
				"code":    APIRequestErrorCode,
				"msg":     APIRequestErrorMsg,
				"content": err.Error(),
			})
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
			// ステータスコードは500にしたいが、レスポンスを返した後なので反映されない
			c.JSON(200, gin.H{
				"code":    APIResposeErrorCode,
				"msg":     APIResposeErrorMsg,
				"content": err.Error(),
			})
			c.Abort()
			return
		}
	}, nil
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
