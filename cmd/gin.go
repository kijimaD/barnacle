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
	if validator, err := barnacle.MakeValidateMiddleware(); err == nil {
		r.Use(validator)
	} else {
		return err
	}

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
