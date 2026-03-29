package yangfen

type BaseRequest struct {
	Uid string `json:"uid" form:"uid"`
}

type BalanceResponse struct {
	Uid     string `json:"uid"`
	Balance int    `json:"balance"`
}

type RechargeRequest struct {
	BaseRequest
	Amount    int   `json:"amount" form:"amount"`
	ExpireSec int64 `json:"expire_sec" form:"expire_sec"`
}

type ConsumeRequest struct {
	BaseRequest
	Amount int `json:"amount" form:"amount"`
}

type TransferRequest struct {
	BaseRequest
	ToUid  string `json:"toUid" form:"toUid"`
	Amount int    `json:"amount" form:"amount"`
}

type RefundRequest struct {
	BaseRequest
	TransactionId string `json:"transactionId" form:"transactionId"`
}

type TransactionRecord struct {
	Id          string `json:"id"`
	Uid         string `json:"uid"`
	Type        string `json:"type"`
	Amount      int    `json:"amount"`
	Balance     int    `json:"balance"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
}

type TransactionListResponse struct {
	List  []TransactionRecord `json:"list"`
	Total int                 `json:"total"`
}

type CommonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
