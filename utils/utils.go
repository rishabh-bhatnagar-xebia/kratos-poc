package utils

import (
	"io"
	"net/http"
)

func SetJson(w http.ResponseWriter) {
	// set content type to JSON
	w.Header().Set("Content-Type", "application/json")
}

func RequestJsonWithCookies(method string, url string, body io.Reader, cookies []*http.Cookie) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return client.Do(req)
}
