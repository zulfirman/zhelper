package model

type Response struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Content interface{} `json:"content"`
	Others  interface{} `json:"others"`
	Path    string      `json:"path"`
}
