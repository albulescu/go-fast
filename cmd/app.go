package main

import (
	"fmt"
	"os"
	"path"

	"github.com/albulescu/go-fast/internal/types"
	"github.com/albulescu/go-fast/pkg/modules/users"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type App struct {
	router  *gin.Engine
	modules []types.AppModule
}

func (app *App) Router() *gin.Engine {
	return app.router
}

func (app *App) RouterHandler(f func(r *types.Request)) func(context *gin.Context) {
	return func(context *gin.Context) {
		f(&types.Request{context})
	}
}

func (app *App) Use(module types.AppModule) {
	module.Setup(app)
	app.modules = append(app.modules, module)
}

func (app *App) Start() {
	app.router.Run()
}

func GetApp() (app *App) {

	app = &App{
		gin.Default(),
		make([]types.AppModule, 0),
	}

	dir, _ := os.Getwd()

	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(
		fmt.Sprintf("%s/config", path.Dir(dir)),
	)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	app.Use(users.GetModule())

	return app
}
