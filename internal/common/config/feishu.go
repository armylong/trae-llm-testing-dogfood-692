package config

// 飞书api
const (
	FeishuApiBitableUrl = "https://open.feishu.cn/open-apis/bitable/v1/apps/%s" // 获取多维表格数据
	// 获取多维表格数据 该接口用于查询数据表中的现有记录，单次最多查询 500 行记录，支持分页获取。 app_token table_id
	// 官方文档: https://open.feishu.cn/document/docs/bitable-v1/app-table-record/search?appId=cli_a94dc0fc84f6dbdd
	FeishuApiDocBaseTablesUrl = "https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records/search"
)
