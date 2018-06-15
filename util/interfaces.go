package util

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"encoding/base64"

	"github.com/kataras/golog"
	"github.com/kataras/iris"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// GetHTTPBody is a method that queries a HTTP endpoint and get the body
func GetHTTPBody(server string, endpoint string) ([]byte, error) {
	headersMap := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
	}
	req, err := CreateHTTPRequest("GET", server+endpoint, headersMap, nil)

	respCode, resp, err := GetHTTPResponse(req)
	if err != nil {
		return nil, err
	}

	if respCode < 200 || respCode >= 400 {
		return nil, errors.New("error while getting HTTP response")
	}
	return resp, nil
}

// CreateHTTPRequest creates a new HTTP request and sets all the necessary headers
func CreateHTTPRequest(method string, url string, headers map[string]string, body []byte) (*http.Request, error) {
	var req *http.Request
	var err error
	if method == "POST" {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else if method == "DELETE" {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	}
	// Checking to see if the request was sucessfully created
	if err != nil {
		return nil, err
	}

	// Setting all the necessary requrest headers
	setRequestHeaders(req, headers)

	return req, nil
}

// GetHTTPResponse gets a Http response for a given request
func GetHTTPResponse(r *http.Request) (int, []byte, error) {
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

// ExtractHTTPBody helper function to extract the HTTP body
func ExtractHTTPBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
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
func BuildResponse(ctx iris.Context, responseData interface{}, isError bool) error {
	if ctx == nil {
		return errors.New("need a iris context")
	}
	SetResponseHeaders(ctx.ResponseWriter(),
		map[string]string{
			"X-Request-ID": fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().Unix())))),
		})
	if isError {
		ctx.StatusCode(iris.StatusInternalServerError)
	} else {
		ctx.StatusCode(iris.StatusOK)
	}
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

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// RandomString return a string of fixed length
func RandomString(length int) string {
	return stringWithCharset(length, charset)
}
