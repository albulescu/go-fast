package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/albulescu/go-fast/internal/types"
	"github.com/albulescu/go-fast/pkg/modules/users"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

type App struct {
	router  *chi.Mux
	db      *gorm.DB
	modules []interface{}
}

func (app *App) Router() *chi.Mux {
	return app.router
}

func (app *App) Database() *gorm.DB {
	return app.db
}

func (app *App) Use(module interface{}) {
	module.(types.AppModule).Setup(app)
	app.modules = append(app.modules, module)
}

func (app *App) HasModule(name string) bool {

	for _, module := range app.modules {
		if module.(types.AppModule).Name() == name {
			return true
		}
	}

	return false
}

func (app *App) GetModule(name string) interface{} {
	for _, module := range app.modules {
		if module.(types.AppModule).Name() == name {
			return module
		}
	}
	return nil
}

func (app *App) Start() {

	for _, module := range app.modules {
		for _, required := range module.(types.AppModule).Requires() {
			if !app.HasModule(required) {
				panic(fmt.Sprintf("Module %s requires %s", module.(types.AppModule).Name(), required))
			}
		}
	}

	// TODO: Refactor here
	port := os.Getenv("PORT")
	fmt.Println("Listen on", port)
	http.ListenAndServe(port, app.router)
}

func GetApp() (app *App) {

	dir, _ := os.Getwd()

	viper.SetDefault("PORT", ":8080")
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(
		fmt.Sprintf("%s/config", path.Dir(dir)),
	)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	db, err := gorm.Open(
		viper.GetString("database.adaptor"),
		viper.GetString("database.args"),
	)

	if err != nil {
		panic(err.Error())
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	app = &App{
		r,
		db,
		make([]interface{}, 0),
	}

	app.Use(users.GetModule())

	return app
}
