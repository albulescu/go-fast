package types

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
)

type App interface {

	// Get router with middlewares from specific domains
	RouterWith(moduleMiddlewares ...string) chi.Router

	// Get router
	Router() *chi.Mux

	// Get database ORM
	Database() *gorm.DB

	// Get a module
	GetModule(name string) interface{}
}

type AppModule interface {
	Setup(app App)
	Name() string
	Middlewares() []func(http.Handler) http.Handler
	Requires() []string
	Run() error
}

type Module struct {
	app App
}

type AppError struct {
	msg      string
	errno    int
	httpcode int
}

func (err *AppError) String() string {
	return err.msg
}

func (err *AppError) ErrNo() int {
	return err.errno
}

func (err *AppError) HttpCode() int {
	return err.httpcode
}

func (err *AppError) Render(w http.ResponseWriter, r *http.Request) {
	render.Status(r, err.HttpCode())
	render.JSON(w, r, render.M{
		"code":    err.ErrNo(),
		"message": err.String(),
	})
}

func NewAppError(msg string, errno, httpcode int) *AppError {
	return &AppError{
		msg:      msg,
		errno:    errno,
		httpcode: httpcode,
	}
}
