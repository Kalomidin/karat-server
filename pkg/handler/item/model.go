package item_handler

type itemRequest struct {
	Name        *string  `json:"name"`
	Kind        *string  `json:"kind"`
	Metal       *string  `json:"metal"`
	Purity      *string  `json:"purity"`
	WeightGrams *float64 `json:"weightGrams"`
	CurrencyID  *string  `json:"currencyId"`
	Cost        *int     `json:"cost"`
	Notes       *string  `json:"notes"`
	CreatedOn   *string  `json:"createdOn"`
	ImageKey    *string  `json:"imageKey"`
	Quantity    *int     `json:"quantity"`
	GroupID     *string  `json:"groupId"`
}

type uploadRequest struct {
	ContentType string `json:"contentType"`
}

type uploadResponse struct {
	UploadURL   string `json:"uploadUrl"`
	ImageKey    string `json:"imageKey"`
	ReadURL     string `json:"readUrl"`
	ContentType string `json:"contentType"`
}

type saleRequest struct {
	SoldOn     *string `json:"soldOn"`
	SoldAmount *int    `json:"soldAmount"`
}
