package uploads

import (
	"net/http"

	"github.com/albulescu/go-fast/internal/core"
	"github.com/albulescu/go-fast/internal/def"
	"github.com/albulescu/go-fast/pkg/modules/users"
	"github.com/defval/inject/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func Setup() def.ModuleFactory {
	return func(di *inject.Container) interface{} {
		return &UploadsModule{}
	}
}

type UploadsModule struct {
	core.AbstractModule
}

func (mod *UploadsModule) GetUsersModule() *users.UsersModule {
	md, _ := mod.App().Module("users")
	usersmod, _ := md.(users.UsersModule)
	return &usersmod
}

func (mod *UploadsModule) Routes(mux *chi.Mux) {
	r := mux.With(mod.GetUsersModule().AuthMiddleware())
	r.Post("/uploads", http.HandlerFunc(mod.upload))
}

func (mod *UploadsModule) upload(w http.ResponseWriter, r *http.Request) {
	user, _ := mod.GetUsersModule().GetUser(w, r)
	render.JSON(w, r, user.GetArbitrary())
}
