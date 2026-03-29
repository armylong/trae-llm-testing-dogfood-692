package cloud_doc

type GetBaseTablesRequest struct {
	AppToken                   string                      `json:"app_token"`                      // 应用凭证 获取方式:表格url上的路径尾部 示例:ZFszben8BaPhvPscIbLcmKsZnYB
	TableID                    string                      `json:"table_id"`                       // 表格ID 获取方式:表格url上的table参数的值 示例: tbluYT98DikJIQp1
	BaseTablesUrlRequestParams *BaseTablesUrlRequestParams `json:"base_tables_url_request_params"` // get参数
	BaseTablesUrlRequestJson   *BaseTablesUrlRequestJson   `json:"base_tables_url_request_json"`   // post参数
}

type GetBaseTablesResponse struct {
}

// 查询参数(拼接到url上面的)
type BaseTablesUrlRequestParams struct {
	UserIDType string `json:"user_id_type"` // 用户ID类型 [非必填] 枚举:[open_id(默认) | union_id | user_id]
	PageToken  string `json:"page_token"`   // 分页标记 [非必填] 第一次请求不填，表示从头开始遍历；分页查询结果还有更多项时会同时返回新的 page_token，下次遍历可采用该 page_token 获取查询结果 示例值："eVQrYzJBNDNONlk4VFZBZVlSdzlKdFJ4bVVHVExENDNKVHoxaVdiVnViQT0="
	PageSize   int    `json:"page_size"`    // 分页大小 [非必填] 最大值为 500 示例值：10 默认值：20
}

// 请求体(json格式)
type BaseTablesUrlRequestJson struct {
	ViewID          string                          `json:"view_id"`          // 视图ID [非必填] 视图ID，多维表格中视图的唯一标识。获取方式：在多维表格的 URL 地址栏中，view_id 参数的值: vew23Yod92
	FieldNames      []string                        `json:"field_names"`      // 字段名称 [非必填] 用于指定本次查询返回记录中包含的字段
	Sort            []BaseTablesUrlRequestJsonSort  `json:"sort"`             // 排序条件 [非必填] 数据校验规则： 长度范围：0 ～ 100
	Filter          *BaseTablesUrlRequestJsonFilter `json:"filter,omitempty"` // 过滤条件 [非必填] 包含条件筛选信息的对象。了解 filter 填写指南和使用示例（如怎样同时使用 and 和 or 逻辑链接词）
	AutomaticFields bool                            `json:"automatic_fields"` // 是否自动计算并返回创建时间（created_time）、修改时间（last_modified_time）、创建人（created_by）、修改人（last_modified_by）这四类字段。默认为 false，表示不返回。示例值：false
}
type BaseTablesUrlRequestJsonSort struct {
	FieldName string `json:"field_name"` // 排序字段的名称 [必填] 示例值："字段1" 数据校验规则： 长度范围：0 字符 ～ 1000 字符
	Desc      bool   `json:"desc"`       // 是否倒序 [必填] 示例值：true | false
}

type BaseTablesUrlRequestJsonFilter struct {
	Conjunction     string                                     `json:"conjunction"`     // 表示条件之间的逻辑连接词 [必填] 示例值："and" 可选值有： and：满足全部条件 or：满足任一条件 数据校验规则： 长度范围：0 字符 ～ 10 字符
	AutomaticFields bool                                       `json:"automaticFields"` // 是否自动计算并返回 [非必填] 创建时间（created_time）、修改时间（last_modified_time）、创建人（created_by）、修改人（last_modified_by）这四类字段。默认为 false，表示不返回。示例值：false
	Conditions      []*BaseTablesUrlRequestJsonFilterCondition `json:"conditions"`      // 筛选条件集合 [非必填] 数据校验规则：长度范围：0 ～ 50
}

type BaseTablesUrlRequestJsonFilterCondition struct {
	FieldName string   `json:"field_name"` // 筛选条件的左值，值为字段的名称 [必填] 示例值："字段1" 数据校验规则： 长度范围：0 字符 ～ 1000 字符
	Operator  string   `json:"operator"`   // 条件运算符 [必填] 示例值："is" 可选值有： is：等于 isNot：不等于（不支持日期字段，了解如何查询日期字段，参考日期字段填写说明） contains：包含（不支持日期字段） doesNotContain：不包含（不支持日期字段） isEmpty：为空 isNotEmpty：不为空 isGreater：大于 isGreaterEqual：大于等于（不支持日期字段） isLess：小于 isLessEqual：小于等于（不支持日期字段） like：LIKE 运算符。暂未支持 in：IN 运算符。暂未支持
	Value     []string `json:"value"`      // 条件的值，可以是单个值或多个值的数组。不同字段类型和不同的 operator 可填的值不同 [必填] 示例值：["文本内容"] 数据校验规则： 长度范围：0 ～ 10
}

type BaseTablesUrlResponse struct {
	Code  int                        `json:"code"` // 错误码，非 0 表示失败
	Msg   string                     `json:"msg"`  // 错误描述
	Data  *BaseTablesUrlResponseData `json:"data,omitempty"`
	Error any                        `json:"error,omitempty"`
}

type BaseTablesUrlResponseData struct {
	Total     int                              `json:"total"`      // 总记录数
	HasMore   bool                             `json:"has_more"`   // 是否还有更多项
	PageToken string                           `json:"page_token"` // 分页标记，当 has_more 为 true 时，会同时返回新的 page_token，否则不返回 page_token
	Items     []*BaseTablesUrlResponseDataItem `json:"items"`      // 记录列表
}

type BaseTablesUrlResponseDataItem struct {
	Fields           map[string]any `json:"fields"`             // 记录字段
	RecordID         string         `json:"record_id"`          // 记录 ID
	CreatedBy        string         `json:"created_by"`         // 创建人
	CreatedTime      int            `json:"created_time"`       // 创建时间
	LastModifiedBy   string         `json:"last_modified_by"`   // 修改人
	LastModifiedTime int            `json:"last_modified_time"` // 最近更新时间
	SharedURL        string         `json:"shared_url"`         // 记录分享链接(批量获取记录接口将返回该字段)
	RecordURL        string         `json:"record_url"`         // 记录链接(检索记录接口将返回该字段)
}
