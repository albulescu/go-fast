package api

import (
	"bitbucket.org/albulescu/remotework/pkg/core"
	"bitbucket.org/albulescu/remotework/pkg/store"
	"github.com/gin-gonic/gin"
)

func Init() {

	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	store.Init()

	r.Use(core.CORS)

	r.Run()
}
