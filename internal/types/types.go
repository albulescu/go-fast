package types

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
)

type App interface {
	Router() *chi.Mux
	Database() *gorm.DB
	GetModule(name string) interface{}
}

type AppModule interface {
	Setup(app App)
	Name() string
	Requires() []string
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
