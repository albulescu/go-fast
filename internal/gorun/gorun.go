package gorun

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/albulescu/go-fast/internal/def"
	"github.com/defval/inject/v2"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

type Config struct {
}

func New(config *Config) def.GoRunApp {

	app := &GoRun{
		mux: chi.NewMux(),
	}

	app.initConfig()

	return app
}

type GoRun struct {
	mux     *chi.Mux
	di      *inject.Container
	modules []interface{}
}

func (app *GoRun) Use(factory def.ModuleFactory) {
	mod := factory(app.Di())
	knd := reflect.ValueOf(mod).MethodByName("SetDi")
	if knd.IsValid() {
		knd.Call([]reflect.Value{
			reflect.ValueOf(app.di),
		})
	}
	app.modules = append(app.modules, mod)
}

func (app *GoRun) initConfig() {
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
}

func (app *GoRun) provideDatabase() (*gorm.DB, func(), error) {

	db, err := gorm.Open(
		viper.GetString("database.adaptor"),
		viper.GetString("database.args"),
	)

	clean := func() {
		db.Close()
	}

	return db, clean, err
}

func (app *GoRun) Di() *inject.Container {

	if app.di == nil {
		app.di = inject.New(
			inject.Provide(app.provideDatabase),
			inject.Provide(func() def.GoRunApp { return app }),
		)
	}

	return app.di
}

type EachModuleFn func(mod interface{})

func (app *GoRun) Module(name string) (interface{}, error) {
	for _, mod := range app.modules {
		str := reflect.TypeOf(mod).Elem().String()
		if strings.Split(str, ".")[0] == name {
			return mod, nil
		}
	}
	return nil, errors.New("Error")
}

func (app *GoRun) Run() {

	for _, mod := range app.modules {
		if m, haveRoutes := mod.(def.ModuleHavingRoutes); haveRoutes {
			m.Routes(app.mux)
		}
	}

	http.ListenAndServe(":8585", app.mux)
}
