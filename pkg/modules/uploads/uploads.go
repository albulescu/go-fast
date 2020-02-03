package uploads

import (
	"net/http"

	"github.com/albulescu/go-fast/internal/types"
	"github.com/go-chi/render"
)

type UploadsModule struct{}

func GetModule() *UploadsModule {
	return &UploadsModule{}
}

func (um *UploadsModule) Setup(app types.App) {

	app.RouterWith("users", "oauth").Get("/uploads", func(w http.ResponseWriter, r *http.Request) {
		render.Data(w, r, []byte("Data"))
	})
}

func (um *UploadsModule) Name() string {
	return "uploads"
}

func (um *UploadsModule) Middlewares() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{}
}

func (um *UploadsModule) Requires() []string {
	return []string{}
}

func (um *UploadsModule) Run() error {
	return nil
}
