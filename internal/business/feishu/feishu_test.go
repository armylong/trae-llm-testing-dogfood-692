package feishu

import (
	"context"
	"fmt"
	"testing"

	feishuLibrary "github.com/armylong/go-library/service/feishu"
)

var ctx = context.Background()

func TestInitUserAccessToken(t *testing.T) {
	code := "bDKkDf7Kyy88Ab18Hfz7KyI20ybzbH4K"
	redirectUri := "https://olfrzjwptnle.ap-northeast-1.clawcloudrun.com"
	userAccessTokenHeader := feishuLibrary.GetUserAccessTokenHeader(code, redirectUri)
	fmt.Println(userAccessTokenHeader)
}

func TestRefreshUserAccessToken(t *testing.T) {
	userAccessTokenHeader := feishuLibrary.GetUserAccessTokenHeader(``, ``)
	fmt.Println(userAccessTokenHeader)
}
