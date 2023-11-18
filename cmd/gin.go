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

// FIXME: ミドルウェア入れたいだけなのに、ハンドラもセットされてるよな…
func initGin(ctx *cli.Context) error {
	r := gin.Default()
	v := barnacle.InitValidator()
	handler := barnacle.InitHandler()
	handler = v.Middleware(handler)
	r.Use(gin.WrapH(handler))
	r.Run()
	return nil
}
