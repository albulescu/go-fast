package main

import "github.com/albulescu/go-fast/pkg/modules/oauth"

func main() {
	app := GetApp()

	app.Use(oauth.GetModule())

	app.Start()
}
