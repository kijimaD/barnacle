package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	mainHandler := initHandler()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mainHandler.ServeHTTP(w, r)
	}))
	defer srv.Close()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "square",
			input:  "/square/2",
			expect: `{"result":100}`,
		},
		{
			name:   "パラメータの型がAPI仕様を満たしていない",
			input:  `/square/"invalid_params"`,
			expect: `{"message":"Bad Request","status":400}`,
		},
		{
			name:   "hello",
			input:  "/hello",
			expect: `{"result":"hello"}`,
		},
		{
			name:   "API仕様に存在しないパス",
			input:  "/not_found",
			expect: `{"message":"Not Found","status":404}`,
		},
		{
			name:   "実装がAPI仕様を満たしていない",
			input:  "/not_impl",
			expect: `{"message":"Internal Server Error","status":500}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := srv.Client().Get(srv.URL + tt.input)
			if err != nil {
				assert.NoError(t, err)
			}
			defer resp.Body.Close()
			contents, err := io.ReadAll(resp.Body)
			if err != nil {
				assert.NoError(t, err)
			}
			if strings.TrimSpace(string(contents)) != tt.expect {
				t.Errorf("got %s want %s", contents, tt.expect)
			}
		})
	}
}
