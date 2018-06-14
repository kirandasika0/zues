package util

// HeadersMap default response headers
var HeadersMap = map[string]string{
	"Content-Type": "application/json",
}

// ZuesRequestBody represent the POST body in a Http request
type ZuesRequestBody struct {
	Data string `json:"data"`
}
