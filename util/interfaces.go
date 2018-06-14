package util

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kataras/iris"
	"encoding/base64"
	"github.com/kataras/golog"
)

// ZuesHTTPService is an interface to abstract all the Http calls made
// from zues to other services
type ZuesHTTPService interface {
	GenerateHttpRequest(server string, endpoint string, headerByKeyValue ...string) *http.Request
	AddHeader(key string, value string) bool
	ExecuteHTTPRequest(r *http.Request) ([]byte, error)
}

// GetHTTPBody is a method that queries a HTTP endpoint and get the body
func GetHTTPBody(server string, endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", server+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Set all the Http headers needed
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// CreateHttpRequest creates a new HTTP request and sets all the necessary headers
func CreateHttpRequest(method string, url string, headers map[string]string, body interface{}) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Setting all the necessary requrest headers
	setRequestHeaders(req, headers)

	return req, nil
}

// GetHttpResponse gets a Http response for a given request
func GetHttpResponse(r *http.Request) (int, []byte, error) {
	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, data, nil

}

func setRequestHeaders(r *http.Request, headers map[string]string) {
	for k, v := range headers {
		r.Header.Set(k, v)
	}

}

// SetResponseHeaders takes a ResponseWriter and HeaderMap and applies them to the writer
func SetResponseHeaders(w http.ResponseWriter, headersMap map[string]string) error {
	if w == nil {
		return errors.New("please provide a ResponseWriter and a HeadersMap")
	}

	if headersMap == nil {
		headersMap = HeadersMap
	} else {
		for k, v := range HeadersMap {
			headersMap[k] = v
		}
	}

	for k, v := range headersMap {
		w.Header().Set(k, v)
	}
	return nil
}

// BuildResponse builds a iris HttpResponse
func BuildResponse(ctx iris.Context, responseData interface{}) error {
	if ctx == nil {
		return errors.New("need a iris context")
	}
	SetResponseHeaders(ctx.ResponseWriter(),
		map[string]string{
			"X-Request-ID": fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().Unix())))),
		})
	ctx.JSON(responseData)
	return nil
}

// Small Helper functions

// EncodeBase64 is a helper to get base64 strings faster
func EncodeBase64(dataToEncode []byte) string {
	return base64.StdEncoding.EncodeToString(dataToEncode)
}

// DecodeBase64 is a helper to get base64 strings faster
func DecodeBase64(dataToDecode string) []byte {
	decodedStr, err := base64.StdEncoding.DecodeString(dataToDecode)
	if err != nil {
		golog.Error("Base64 decoding failed")
		return []byte("")
	}
	return decodedStr
}