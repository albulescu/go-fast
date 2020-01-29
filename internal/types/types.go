package types

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Role byte
type presence uint8

const (
	OFFLINE presence = iota
	ONLINE
)

const (
	Driver Role = 1 << (iota + 1)
	Helper
)

type dstring string

func (ds *dstring) String() string {
	return ""
}

// User type
type User struct {
	ID             string    `json:"_id" bson:"_id"`
	FirstName      string    `json:"firstName" bson:"firstName"`
	LastName       string    `json:"lastName" bson:"lastName"`
	Username       string    `json:"username" bson:"username"`
	Password       string    `json:"password,omitempty" bson:"password,omitempty"`
	Email          string    `json:"emails" bson:"emails"`
	Role           Role      `json:"role" bson:"role"`
	Timezone       string    `json:"timezone" bson:"timezone,omitempty"`
	Avatar         string    `json:"avatar" bson:"avatar,omitempty"`
	Activated      bool      `json:"activated" bson:"activated"`
	RegisteredDate time.Time `json:"registered_date" bson:"registered_date"`
}

type Car struct {
	ID    string    `json:"_id" bson:"_id"`
	Label string    `json:"label" bson:"label"`
	Name  string    `json:"name" bson:"name"`
	Model string    `json:"model" bson:"model"`
	Year  time.Time `json:"year" bson:"year"`
}

type Service struct {
	ID   string  `json:"_id" bson:"_id"`
	Icon string  `json:"icon" bson:"icon"`
	Name dstring `json:"name" bson:"name"`
}

type UserApp struct {
	ID          string    `json:"_id" bson:"_id"`
	User        string    `json:"user" bson:"user"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	Domain      string    `json:"domain" bson:"domain"`
	Secret      string    `json:"secret" bson:"secret"`
	Enabled     bool      `json:"enabled" bson:"enabled"`
	LastUsed    time.Time `json:"last_used,omitempty" bson:"last_used,omitempty"`
	Date        time.Time `json:"date" bson:"date"`
}

func (app *UserApp) GetID() string {
	return app.ID
}

func (app *UserApp) GetSecret() string {
	return app.Secret
}

func (app *UserApp) GetDomain() string {
	return app.Domain
}

func (app *UserApp) GetUserID() string {
	return app.User
}

type Request struct {
	Context *gin.Context
}

type App interface {
	Router() *gin.Engine
	RouterHandler(f func(r *Request)) func(c *gin.Context)
}

type AppModule interface {
	Setup(app App)
}

type Module struct {
	app App
}
