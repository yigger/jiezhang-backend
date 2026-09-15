package types

type PayeeWriteRequest struct {
	Payee struct {
		Name string `json:"name"`
	} `json:"payee"`
	Name string `json:"name"`
}

type PayeeListItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
