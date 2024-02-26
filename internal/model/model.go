package model

type Child struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Tokenize struct {
	ID   string  `json:"id"`
	Data []Child `json:"data"`
}

type TokenizeResponse struct {
	ID   string  `json:"id"`
	Data []Child `json:"data"`
}

type Detokenize struct {
	ID   string  `json:"id"`
	Data []Child `json:"data"`
}

type DetokenizeResponse struct {
	ID   string          `json:"id"`
	Data []*ChildReceipt `json:"data"`
}

type ChildReceipt struct {
	Key   string     `json:"key"`
	Value *ChildResp `json:"value"`
}

type ChildResp struct {
	Found bool   `json:"found"`
	Datum string `json:"datum"`
}

type All struct {
	Tokens []*Tokenize `json:"tokens"`
}

type Resp interface {
}

type Response struct {
	Resp  `json:"resp"`
	Code  int      `json:"code"`
	Error []string `json:"error"`
}
