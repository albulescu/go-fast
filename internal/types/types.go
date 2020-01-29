package types

import (
	"github.com/gin-gonic/gin"
)

type Request struct {
	Context *gin.Context
}

type App interface {
	Router() *gin.Engine
}

type AppModule interface {
	Setup(app App)
}

type Module struct {
	app App
}
