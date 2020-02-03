package users

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/albulescu/go-fast/internal/types"
	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
	"github.com/volatiletech/authboss"
	abclientstate "github.com/volatiletech/authboss-clientstate"
	_ "github.com/volatiletech/authboss/auth"
	"github.com/volatiletech/authboss/confirm"
	"github.com/volatiletech/authboss/defaults"
	"github.com/volatiletech/authboss/lock"
	aboauth "github.com/volatiletech/authboss/oauth2"
	"github.com/volatiletech/authboss/otp/twofactor"
	"github.com/volatiletech/authboss/otp/twofactor/sms2fa"
	"github.com/volatiletech/authboss/otp/twofactor/totp2fa"
	_ "github.com/volatiletech/authboss/register"
	"github.com/volatiletech/authboss/remember"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

type UsersModule struct {
	app  types.App
	auth *authboss.Authboss
}

func GetModule() types.AppModule {
	return &UsersModule{}
}

func (u *UsersModule) GetAuthboss() *authboss.Authboss {
	return u.auth
}

func (u *UsersModule) Setup(app types.App) {
	u.app = app
	app.Database().AutoMigrate(&User{})

	isApi := true
	sessionCookieName := viper.GetString("auth.cookie.name")
	cookieStoreKey := viper.GetString("auth.cookie.key")
	sessionStoreKey := viper.GetString("auth.session.key")
	cookieStore := abclientstate.NewCookieStorer([]byte(cookieStoreKey), nil)
	cookieStore.HTTPOnly = false
	cookieStore.Secure = false
	sessionStore := abclientstate.NewSessionStorer(sessionCookieName, []byte(sessionStoreKey), nil)
	cstore := sessionStore.Store.(*sessions.CookieStore)
	cstore.Options.HttpOnly = false
	cstore.Options.Secure = false

	duration, err := time.ParseDuration(viper.GetString("auth.cookie.maxage"))
	if err != nil {
		panic(
			fmt.Sprintf("Invalid auth.cookie.maxage: %s", viper.GetString("auth.cookie.maxage")),
		)
	}
	cstore.MaxAge(int(duration.Seconds()))

	// Configure authboss
	ab := authboss.New()
	u.auth = ab
	ab.Config.Storage.SessionState = sessionStore
	ab.Config.Storage.CookieState = cookieStore
	ab.Config.Paths.RootURL = fmt.Sprintf("http://localhost%s", viper.GetString("http.port"))
	ab.Config.Storage.Server = NewUserStore(app.Database())
	ab.Config.Modules.TwoFactorEmailAuthRequired = true
	ab.Config.Modules.RegisterPreserveFields = []string{"email", "name"}
	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	ab.Config.Core.MailRenderer = &MailRenderer{}
	ab.Config.Core.Logger = defaults.NewLogger(os.Stdout)
	ab.Config.Core.Responder = defaults.NewResponder(&MailRenderer{})
	defaults.SetCore(&ab.Config, isApi, false)

	emailRule := defaults.Rules{
		FieldName: "email", Required: true,
		MatchError: "Must be a valid e-mail address",
		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]+`),
	}
	passwordRule := defaults.Rules{
		FieldName: "password", Required: true,
		MinLength: 4,
	}
	nameRule := defaults.Rules{
		FieldName: "name", Required: true,
		MinLength: 2,
	}

	ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
		ReadJSON: true,
		Rulesets: map[string][]defaults.Rules{
			"register":    {emailRule, passwordRule, nameRule},
			"recover_end": {passwordRule},
		},
		Confirms: map[string][]string{
			"register":    {"password", authboss.ConfirmPrefix + "password"},
			"recover_end": {"password", authboss.ConfirmPrefix + "password"},
		},
		Whitelist: map[string][]string{
			"register": {"email", "name", "password"},
		},
	}

	// Set up 2fa
	twofaRecovery := &twofactor.Recovery{Authboss: ab}
	if err := twofaRecovery.Setup(); err != nil {
		panic(err)
	}

	totp := &totp2fa.TOTP{Authboss: ab}
	if err := totp.Setup(); err != nil {
		panic(err)
	}

	sms := &sms2fa.SMS{Authboss: ab, Sender: smsLogSender{}}
	if err := sms.Setup(); err != nil {
		panic(err)
	}
	ab.Config.Modules.OAuth2Providers = map[string]authboss.OAuth2Provider{
		"google": {
			OAuth2Config: &oauth2.Config{
				ClientID:     viper.GetString("auth.modules.oauth.google.client_id"),
				ClientSecret: viper.GetString("auth.modules.oauth.google.client_secret"),
				Scopes:       []string{`profile`, `email`},
				Endpoint:     google.Endpoint,
			},
			FindUserDetails: aboauth.GoogleUserDetails,
		},
		"facebook": {
			OAuth2Config: &oauth2.Config{
				ClientID:     viper.GetString("auth.modules.oauth.facebook.client_id"),
				ClientSecret: viper.GetString("auth.modules.oauth.facebook.client_secret"),
				Scopes:       []string{`name`, `email`},
				Endpoint:     facebook.Endpoint,
			},
			FindUserDetails: aboauth.FacebookUserDetails,
		},
	}

	// Initialize authboss (instantiate modules etc.)
	if err := ab.Init(); err != nil {
		panic(err)
	}
}
func (u *UsersModule) Run() error {

	// Routes
	middlewares := []func(http.Handler) http.Handler{
		authboss.ModuleListMiddleware(u.auth),
		u.auth.LoadClientStateMiddleware,
		remember.Middleware(u.auth),
		u.dataInjector,
	}

	u.app.Router().With(middlewares...).Mount("/auth", http.StripPrefix("/auth", u.auth.Config.Core.Router))

	return nil
}
func (u *UsersModule) Middlewares() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		authboss.Middleware2(
			u.auth,
			authboss.RequireNone,
			authboss.RespondUnauthorized,
		),
		lock.Middleware(u.auth),
		confirm.Middleware(u.auth),
	}
}

func (u *UsersModule) Name() string {
	return "users"
}

func (u *UsersModule) Requires() []string {
	return []string{}
}

func (module *UsersModule) Auth(email, password string) (*User, error) {

	user := User{}
	query := module.app.Database().Where("email = ?", email).Find(&user).Limit(1)

	var count int

	if query.Count(&count); count == 0 {
		return nil, authboss.ErrUserNotFound
	}

	return &user, bcrypt.CompareHashAndPassword(
		[]byte(user.GetPassword()),
		[]byte(password),
	)
}

type smsLogSender struct {
}

// Send an SMS
func (s smsLogSender) Send(_ context.Context, number, text string) error {
	fmt.Println("sms sent to:", number, "contents:", text)
	return nil
}

func (u *UsersModule) dataInjector(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := u.layoutData(w, &r)
		r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
		handler.ServeHTTP(w, r)
	})
}

func (u *UsersModule) layoutData(w http.ResponseWriter, r **http.Request) authboss.HTMLData {
	currentUserName := ""
	userInter, err := u.auth.LoadCurrentUser(r)
	if userInter != nil && err == nil {
		currentUserName = userInter.(*User).Name
	}

	return authboss.HTMLData{
		"loggedin":          userInter != nil,
		"current_user_name": currentUserName,
		"flash_success":     authboss.FlashSuccess(w, *r),
		"flash_error":       authboss.FlashError(w, *r),
	}
}

type MailRenderer struct{}

func (mr *MailRenderer) Load(names ...string) error {
	fmt.Println("Load mail templates:", names)
	return nil
}

// Render the given template
func (mr *MailRenderer) Render(ctx context.Context, page string, data authboss.HTMLData) (output []byte, contentType string, err error) {
	fmt.Println("Render mail template:", page, data)
	return []byte("Mail"), "text/html", nil
}
