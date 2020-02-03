package main

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	app := GetApp()
	// app.Use(oauth.GetModule())
	app.Start()
}
