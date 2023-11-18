package cmd

import (
	"barnacle/barnacle"

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

func initGin(ctx *cli.Context) error {
	r := gin.Default()
	v := barnacle.InitValidator()
	mainHandler := v.Middleware(barnacle.InitHandler())
	r.Use(gin.WrapH(mainHandler))
	r.Run()
	return nil
}
