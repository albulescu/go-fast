package def

import (
	"github.com/apsdehal/go-logger"
	"github.com/go-chi/chi"
)

type GoRunApp interface {
	Use(factory ModuleFactory)
	Module(name string) (interface{}, error)
	Log() *logger.Logger
	Run()
}

type ModuleHavingRoutes interface {
	Routes(mux *chi.Mux)
}

type Mailer interface {
	Send(template string, params map[string]interface{}, subject, to string) error
}
