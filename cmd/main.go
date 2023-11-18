package cmd

import (
	"github.com/urfave/cli/v2"
)

func NewApp() *cli.App {
	app := &cli.App{}
	app.Name = "barnacle"
	app.Usage = "server"
	app.Description = "server"
	app.Version = "v0.0.0"
	app.EnableBashCompletion = true
	app.DefaultCommand = CmdStd.Name
	app.Commands = []*cli.Command{
		CmdGin,
		CmdStd,
	}
	return app
}

func RunApp(app *cli.App, args ...string) error {
	err := app.Run(args)
	if err != nil {
		return err
	}

	return err
}
