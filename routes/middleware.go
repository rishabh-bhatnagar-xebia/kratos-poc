package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"kratos-rbac/constants"
	"kratos-rbac/model"
	"net/http"
	"strings"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func makeRequest(url string, cookies string, method string) string {
	// Create a new GET request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}

	// Set cookies in the request header
	req.Header.Set("Cookie", cookies)

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Print the response status code and body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request)
		return
		// set the cookies on the ory client
		var cookies string

		cookies = request.Header.Get("Cookie")
		fmt.Println("cookies", cookies)

		// handle 401
		content := makeRequest(constants.KratosPublicApi+"sessions/whoami", cookies, http.MethodGet)
		if strings.Contains(content, "401") {
			fmt.Println(content)
			http.Error(writer, "unauthorized\n", 401)
			return
		}

		var si model.SessionInfo
		err := json.NewDecoder(strings.NewReader(content)).Decode(&si)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err = si.Validate(); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if !si.OfRole(model.ADMIN) {
			// todo: refactor error message if we're giving away more information
			http.Error(writer, fmt.Sprintf("only %s can access this endpoint", model.ADMIN), 401)
			return
		}

		next.ServeHTTP(writer, request)
		return
	}
}

func RequestLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("got a %s request\n", r.Method)
		//w.Write([]byte(fmt.Sprintf("received a %s request\n", r.Method)))
		next.ServeHTTP(w, r)
	}
}
