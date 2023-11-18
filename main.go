package main

import (
	"barnacle/cmd"
	"os"
)

func main() {
	app := cmd.NewApp()
	_ = cmd.RunApp(app, os.Args...)
}
