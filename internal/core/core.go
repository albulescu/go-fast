package core

import (
	"net/http"

	"github.com/albulescu/go-fast/internal/def"
	"github.com/defval/inject/v2"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
)

type AbstractModule struct {
	di *inject.Container
}

func (am *AbstractModule) SetDi(di *inject.Container) {
	am.di = di
}

func (am *AbstractModule) Di() *inject.Container {
	return am.di
}

func (am *AbstractModule) Emit(event string) {

}

func (am *AbstractModule) Error(w http.ResponseWriter, r *http.Request, status, code int, message string) {
	render.Status(r, status)
	render.JSON(w, r, map[string]interface{}{
		"code":    code,
		"message": message,
	})
}

func (am *AbstractModule) Db() *gorm.DB {
	var db *gorm.DB
	am.di.Extract(&db)
	return db
}

func (am *AbstractModule) Mailer() def.Mailer {
	impl, err := am.App().Module("mailer")
	if err != nil {
		panic(err)
	}
	return impl.(def.Mailer)
}

func (am *AbstractModule) App() def.GoRunApp {
	var app def.GoRunApp
	am.di.Extract(&app)
	return app
}
