package main

import (
	"github.com/albulescu/go-fast/internal/gorun"
	"github.com/albulescu/go-fast/pkg/modules/mailer"
	"github.com/albulescu/go-fast/pkg/modules/profile"
	"github.com/albulescu/go-fast/pkg/modules/uploads"
	"github.com/albulescu/go-fast/pkg/modules/users"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	app := gorun.New(&gorun.Config{})
	app.Use(users.Setup(&users.Config{}))
	app.Use(mailer.Setup())
	app.Use(uploads.Setup())
	app.Use(profile.Setup())
	app.Run()
}
