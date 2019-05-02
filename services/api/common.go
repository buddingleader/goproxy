package api

// Result for http response
type Result struct {
	OK      bool        `json:"ok"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}
