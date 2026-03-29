package cloud_doc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/armylong/armylong-go/internal/common/config"
	cloudDocCs "github.com/armylong/armylong-go/internal/cs/feishu/cloud_doc"
	feishuLibrary "github.com/armylong/go-library/service/feishu"
	"github.com/armylong/go-library/service/httpx"
	"github.com/spf13/cast"
)

type baseTablesBusiness struct{}

var BaseTablesBusiness = &baseTablesBusiness{}

func (b *baseTablesBusiness) GetBaseTables(ctx context.Context, req *cloudDocCs.GetBaseTablesRequest) (*cloudDocCs.BaseTablesUrlResponse, error) {
	if req == nil || req.AppToken == "" || req.TableID == "" {
		return nil, fmt.Errorf("req is nil or app_token or table_id is empty")
	}
	headerAuthorization := feishuLibrary.GetUserAccessTokenHeader(``, ``)
	apiUrl := fmt.Sprintf(config.FeishuApiDocBaseTablesUrl, req.AppToken, req.TableID)
	// 拼接参数
	if req.BaseTablesUrlRequestParams != nil {
		params := url.Values{}
		if req.BaseTablesUrlRequestParams.UserIDType != "" {
			params.Add("user_id_type", req.BaseTablesUrlRequestParams.UserIDType)
		}
		if req.BaseTablesUrlRequestParams.PageSize > 0 {
			params.Add("page_size", cast.ToString(req.BaseTablesUrlRequestParams.PageSize))
		}
		if req.BaseTablesUrlRequestParams.PageToken != "" {
			params.Add("page_token", req.BaseTablesUrlRequestParams.PageToken)
		}
		if len(params) > 0 {
			apiUrl += "?" + params.Encode()
		}
	}

	postReq := &cloudDocCs.BaseTablesUrlRequestJson{}
	postReqByte, _ := json.Marshal(postReq)
	urlResponseByte, err := httpx.PostWithHeader(apiUrl, postReqByte, map[string]string{
		`Authorization`: headerAuthorization,
		`Content-Type`:  `application/json`,
	})
	if err != nil {
		return nil, errors.New(`请求异常`)
	}

	// 处理响应
	urlResponse := &cloudDocCs.BaseTablesUrlResponse{}
	err = json.Unmarshal(urlResponseByte, &urlResponse)
	if err != nil {
		return nil, err
	}
	if urlResponse.Code != 0 {
		return nil, fmt.Errorf("code is not 0, msg is %s", urlResponse.Msg)
	}

	return urlResponse, err
}
