package cmd

import (
	"barnacle/barnacle"
	"net/http"

	"github.com/urfave/cli/v2"
)

var CmdStd = &cli.Command{
	Name:        "std",
	Usage:       "標準サーバを起動する",
	Description: "標準サーバを起動する",
	Action:      initStd,
	Flags:       []cli.Flag{},
}

func initStd(ctx *cli.Context) error {
	v := barnacle.InitValidator()
	mainHandler := v.Middleware(barnacle.InitHandler())
	http.ListenAndServe(":8080", mainHandler)
	return nil
}
