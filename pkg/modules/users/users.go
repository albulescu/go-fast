package users

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/albulescu/go-fast/internal/core"
	"github.com/albulescu/go-fast/internal/def"
	"github.com/defval/inject/v2"
	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound         = errors.New("User not found")
	ErrInvalidPassword      = errors.New("Invalid password")
	ErrUserNotActivated     = errors.New("User not activated")
	ErrAuthorizationExpired = errors.New("Access token expired")
)

type User struct {
	ID uint `gorm:"primary_key"`

	Email    string `gorm:"unique;not null"`
	Password string

	Name string

	ActivationCode string `gorm:"unique"`
	Activated      bool

	CreatedAt time.Time
}

func (u *User) GetArbitrary() map[string]string {
	return map[string]string{
		"id":   fmt.Sprint(u.ID),
		"name": u.Name,
	}
}

type UserRegister struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	VerifyPassword string `json:"verify_password" validate:"required,eqfield=Password"`
	Name           string `json:"name" validate:"required"`
}

type UserAuth struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Config struct{}

func Setup(config *Config) def.ModuleFactory {
	return func(di *inject.Container) interface{} {
		return &UsersModule{}
	}
}

type UsersModule struct {
	core.AbstractModule
}

func (mod *UsersModule) Routes(mux *chi.Mux) {

	mux.Post("/register", http.HandlerFunc(mod.register))
	mux.Post("/register/activate", http.HandlerFunc(mod.registerActivate))

	mux.Post("/auth", http.HandlerFunc(mod.auth))

	r := mux.With(mod.AuthMiddleware())
	r.Get("/me", http.HandlerFunc(mod.me))
	r.Get("/me/avatar", http.HandlerFunc(mod.avatar))
}

func (mod *UsersModule) avatar(w http.ResponseWriter, r *http.Request) {
	render.Data(w, r, []byte("Avatar bytes"))
}

func (mod *UsersModule) auth(w http.ResponseWriter, r *http.Request) {

	auth := UserAuth{}

	if err := render.DecodeJSON(r.Body, &auth); err != nil {
		mod.Error(w, r, 400, 1000, "Failed to read json payload")
		return
	}

	validate := validator.New()
	if err := validate.Struct(auth); err != nil {
		mod.Error(w, r, 400, 1000, err.Error())
		return
	}

	user, err := mod.Auth(auth.Email, auth.Password)

	if err != nil {
		mod.Error(w, r, http.StatusUnauthorized, 1000, err.Error())
		return
	}

	access, err := mod.CreateJWT(user)

	if err != nil {
		mod.Error(w, r, http.StatusInternalServerError, 1000, err.Error())
		return
	}

	render.JSON(w, r, def.M{
		"access_token": access,
	})
}

//
// Register new user
//
func (mod *UsersModule) register(w http.ResponseWriter, r *http.Request) {

	register := UserRegister{}

	if err := render.DecodeJSON(r.Body, &register); err != nil {
		mod.Error(w, r, 400, 1000, "Failed to read json payload")
		return
	}

	validate := validator.New()
	if err := validate.Struct(register); err != nil {
		mod.Error(w, r, 400, 1000, err.Error())
		return
	}

	password, err := bcrypt.GenerateFromPassword(
		[]byte(register.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		mod.Error(w, r, 400, 1000, "Failed to compute password hash")
		return
	}

	user := &User{
		Email:     register.Email,
		Password:  string(password),
		Name:      register.Name,
		CreatedAt: time.Now(),
	}

	db := mod.Db()
	db.AutoMigrate(&user)

	err = db.Create(&user).Error
	if err != nil {
		switch err.(*pq.Error).Code {
		case "23505":
			mod.Error(w, r, 400, 1000, "User with email already exists")
		}
		return
	}

	activationCode := xid.New().String()

	mod.Db().Model(&user).Update("activation_code", activationCode)

	mod.Mailer().Send(
		"register_activation",
		map[string]interface{}{
			"NAME": user.Name,
			"CODE": activationCode,
		},
		"{{.NAME}}, thanks for registering. Please activate your account.",
		user.Email,
	)

	w.WriteHeader(201)
}

// Register new user route
func (mod *UsersModule) registerActivate(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")
	db := mod.Db()
	user := &User{}

	if db.Where("activation_code = ?", code).First(user).RecordNotFound() {
		mod.Error(w, r, 400, 1000, "Invalid activation code")
		return
	}

	db.Model(&user).Updates(
		def.M{
			"activation_code": "",
			"activated":       true,
		},
	)

	if db.Error != nil {
		mod.Error(w, r, 400, 1000, "Failed to activate account")
		return
	}

	mod.Mailer().Send(
		"register_activation_complete",
		map[string]interface{}{
			"NAME": user.Name,
		},
		"{{.NAME}}, your account is activated.",
		user.Email,
	)
}

func (mod *UsersModule) me(w http.ResponseWriter, r *http.Request) {
	if user, err := mod.GetUser(w, r); err == nil {
		render.JSON(w, r, user.GetArbitrary())
	}
}

func (mod *UsersModule) DecodeAuthorization(r *http.Request) (*User, error) {

	authorization := r.Header.Get("Authorization")

	if authorization == "" {
		return nil, errors.New("No authorization header provided")
	}

	access := strings.Split(authorization, "Bearer ")

	if len(access) != 2 {
		return nil, errors.New("Invalid authorization")
	}

	id, data, err := jose.Decode(access[1], []byte(viper.GetString("auth.key")))

	eatv, err := strconv.ParseInt(data["eat"].(string), 10, 64)
	eat := time.Unix(eatv, 0)

	if eat.Before(time.Now()) {
		return nil, ErrAuthorizationExpired
	}

	if err != nil {
		return nil, errors.New("Failed to decode access_token")
	}

	db := mod.Db()
	user := &User{}

	if db.Where("id = ?", id).First(user).RecordNotFound() {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (mod *UsersModule) GetUser(w http.ResponseWriter, r *http.Request) (*User, error) {

	user, err := mod.DecodeAuthorization(r)

	if err != nil {
		mod.Error(w, r, 401, 1000, errors.Wrap(err, "Unauthorized").Error())
		return nil, err
	}

	return user, err
}

// Authenticate user
func (mod *UsersModule) Auth(email, password string) (*User, error) {

	db := mod.Db()
	user := &User{}

	if db.Where("email = ?", email).First(user).RecordNotFound() {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	if !user.Activated {
		return nil, ErrUserNotActivated
	}

	return user, nil
}

func (mod *UsersModule) CreateJWT(user *User) (string, error) {

	issued := jose.Header("iat", strconv.FormatInt(user.CreatedAt.Unix(), 10))
	aeat := strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)

	return jose.Sign(
		fmt.Sprint(user.ID),
		jose.HS256,
		[]byte(viper.GetString("auth.key")),
		issued,
		jose.Header("eat", aeat),
		jose.Header("scope", "app"),
	)
}

func (mod *UsersModule) AuthMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := mod.GetUser(w, r); err == nil {
				h.ServeHTTP(w, r)
			}
		})
	}
}
