package oauth

import "github.com/albulescu/go-fast/internal/types"

type OAuthModule struct {
}

func (oauth *OAuthModule) Setup(app types.App) {

}

func GetModule() types.AppModule {
	return &OAuthModule{}
}
