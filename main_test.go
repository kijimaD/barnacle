package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	mainHandler := initHandler()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 先にhttptest.ServerのURLを確定させるため
		// 外部から処理を注入できるようにする。変数mainHandlerに値を入れると、srvの動きが変わる
		mainHandler.ServeHTTP(w, r)
	}))
	defer srv.Close()

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
	printResp(srv.Client().Get(srv.URL + "/hello"))     // 処理はsquareHandler...
	printResp(srv.Client().Get(srv.URL + "/not_found")) // error
}
