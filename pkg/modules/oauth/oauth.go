package oauth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/albulescu/go-fast/internal/types"
	"github.com/albulescu/go-fast/pkg/modules/users"
	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"github.com/volatiletech/authboss"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	oauth2Store "gopkg.in/oauth2.v3/store"
)

type OauthModule struct {
	app types.App
}

func GetModule() types.AppModule {
	return &OauthModule{}
}

func (module *OauthModule) UsersModule() *users.UsersModule {

	users, exists := module.app.GetModule("users").(*users.UsersModule)

	if !exists {
		panic("Users module does not exist")
	}

	return users
}

func (module *OauthModule) GetAuthboss() *authboss.Authboss {
	return module.UsersModule().GetAuthboss()
}

func (module *OauthModule) createApp(w http.ResponseWriter, r *http.Request) {

	user, _ := module.GetAuthboss().CurrentUser(r)

	app := UserApp{
		User: user.(*users.User),
	}

	err := render.DecodeJSON(r.Body, &app)
	if err != nil {
		render.Data(w, r, []byte(err.Error()))
		render.Status(r, 400)
		return
	}

	db := module.app.Database()
	db.Create(app).Scan(&app)

	if db.Error != nil {
		render.Data(w, r, []byte(db.Error.Error()))
		render.Status(r, 500)
		return
	}

	render.Status(r, 201)
}

func (module *OauthModule) Setup(app types.App) {
	module.app = app
	app.Database().AutoMigrate(&UserApp{})
	usersModule, valid := app.GetModule("users").(*users.UsersModule)
	if !valid {
		panic("Not user module")
	}
	router := usersModule.AuthRoute()
	router.Post("/apps", module.createApp)

	cliStorage := &OauthClientsStore{db: app.Database()}

	manager := manage.NewManager()
	tokenStorage, _ := oauth2Store.NewMemoryTokenStore()
	// manager.MapTokenModel(models.NewToken())
	manager.SetAuthorizeCodeExp(time.Minute * 10)
	manager.MapAuthorizeGenerate(generates.NewAuthorizeGenerate())
	manager.MapTokenStorage(tokenStorage)
	manager.MapClientStorage(cliStorage)
	manager.MapAccessGenerate(&TokenGenerate{})

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)

	srv.SetClientInfoHandler(func(r *http.Request) (clientID, clientSecret string, err error) {

		clientID = r.FormValue("client_id")

		if clientID == "" {
			return "", "", errors.New("No client defined")
		}

		cli, errSto := cliStorage.GetByID(clientID)

		if errSto != nil {
			return "", "", errSto
		}

		return cli.GetID(), cli.GetSecret(), nil
	})

	srv.SetPasswordAuthorizationHandler(func(email, password string) (userID string, err error) {
		usrs := app.GetModule("users")
		user, err := usrs.(*users.UsersModule).Auth(email, password)
		return fmt.Sprint(user.ID), err
	})

	app.Router().Post("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		srv.HandleTokenRequest(w, r)
	})

	app.Router().Get("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		srv.HandleAuthorizeRequest(w, r)
	})
}

func (module *OauthModule) Name() string {
	return "oauth"
}

func (module *OauthModule) Requires() []string {
	return []string{"users"}
}

type OauthClientsStore struct {
	db *gorm.DB
}

func (cs *OauthClientsStore) GetByID(id string) (cli oauth2.ClientInfo, err error) {

	if id == "main" {
		return &UserApp{
			ID:     0,
			Date:   time.Now(),
			Domain: "https://localhost:8080",
			Secret: "secret",
			User:   nil,
		}, err
	}

	app := UserApp{}
	cs.db.Where("id = ?", id).Find(&app)

	return &app, nil
}

type UserApp struct {
	ID          uint        `gorm:"primary_key"`
	User        *users.User `json:"-"`
	Name        string
	Description string
	Domain      string
	Secret      string
	Enabled     bool
	LastUsed    time.Time
	Date        time.Time
}

func (uapp *UserApp) GetID() string {
	return fmt.Sprint(uapp.ID)
}

func (uapp *UserApp) GetSecret() string {
	return uapp.Secret
}

func (uapp *UserApp) GetDomain() string {
	return uapp.Domain
}

func (uapp *UserApp) GetUserID() string {
	return fmt.Sprint(uapp.User.ID)
}

type TokenGenerate struct{}

func (tg *TokenGenerate) Token(data *oauth2.GenerateBasic, isGenRefresh bool) (access, refresh string, err error) {

	issued := jose.Header("iat", strconv.FormatInt(data.CreateAt.Unix(), 10))
	aeat := strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)

	access, err = jose.Sign(
		data.UserID,
		jose.HS256,
		[]byte(data.Client.GetSecret()),
		issued,
		jose.Header("client", data.Client.GetID()),
		jose.Header("eat", aeat),
		jose.Header("scope", data.TokenInfo.GetScope()),
	)

	if isGenRefresh {
		refresh, err = jose.Sign(
			data.UserID,
			jose.HS256,
			[]byte(data.Client.GetSecret()),
			issued,
			jose.Header("client", data.Client.GetID()),
			jose.Header("eat", time.Now().Add(data.TokenInfo.GetRefreshExpiresIn()).Unix()),
			jose.Header("scope", "refresh"),
		)
	}

	return
}

type TokenRequest struct {
	GrantType   string `json:"grant_type"`
	RedirectURI string `json:"redirect_uri"`
	ClientID    string `json:"client_id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Scope       string `json:"scope"`
	State       string `json:"state"`
}
