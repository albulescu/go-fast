package uploads

import "github.com/albulescu/go-fast/internal/types"

type UploadsModule struct {
}

func (oauth *UploadsModule) Setup(app types.App) {

}

func GetModule() types.AppModule {
	return &UploadsModule{}
}
