package profile

import (
	"github.com/albulescu/go-fast/internal/core"
	"github.com/albulescu/go-fast/internal/def"
	"github.com/defval/inject/v2"
)

func Setup() def.ModuleFactory {
	return func(di *inject.Container) interface{} {
		return &ProfileModule{}
	}
}

type ProfileModule struct {
	core.AbstractModule
}
