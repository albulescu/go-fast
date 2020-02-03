package def

import "github.com/defval/inject/v2"

type M map[string]interface{}

type ModuleFactory func(di *inject.Container) interface{}
