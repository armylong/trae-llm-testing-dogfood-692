package cloud_doc

import (
	"context"
	"testing"

	cloudDocCs "github.com/armylong/armylong-go/internal/cs/feishu/cloud_doc"
)

var ctx = context.Background()

func Test_baseTablesBusiness_GetBaseTables(t *testing.T) {
	BaseTablesBusiness.GetBaseTables(ctx, &cloudDocCs.GetBaseTablesRequest{
		AppToken: "ZFszben8BaPhvPscIbLcmKsZnYB",
		TableID:  "tbluYT98DikJIQp1",
		BaseTablesUrlRequestParams: &cloudDocCs.BaseTablesUrlRequestParams{
			PageSize: 1,
		},
	})
}
