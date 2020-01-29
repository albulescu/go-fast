package users

import "github.com/albulescu/go-fast/internal/types"

import "github.com/gin-gonic/gin"

type UsersModule struct {
}

func (u *UsersModule) Setup(app types.App) {

	app.Router().POST("users", func(ctx *gin.Context) {
		ctx.Data(200, "text/plain", []byte("Test"))
	})

	app.Router().GET("users", func(ctx *gin.Context) {
		ctx.Data(200, "text/plain", []byte("Test"))
	})

	app.Router().GET("users/:id", func(ctx *gin.Context) {
		ctx.Data(200, "text/plain", []byte("Test"))
	})

	app.Router().PUT("users/:id", func(ctx *gin.Context) {
		ctx.Data(200, "text/plain", []byte("Test"))
	})
}

func GetModule() types.AppModule {
	return &UsersModule{}
}
