package feishu

import (
	"context"

	feishuBusiness "github.com/armylong/armylong-go/internal/business/feishu/cloud_doc"
	cloudDocCs "github.com/armylong/armylong-go/internal/cs/feishu/cloud_doc"
)

type FeishuCloudDocController struct{}

func (c *FeishuCloudDocController) ActionGetBaseTables(ctx context.Context, req *cloudDocCs.GetBaseTablesRequest) (*cloudDocCs.BaseTablesUrlResponse, error) {
	return feishuBusiness.BaseTablesBusiness.GetBaseTables(ctx, req)
}
