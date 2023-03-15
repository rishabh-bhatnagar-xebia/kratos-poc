package routes

import (
	"net/http"
)

func endpoint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		CreateUser(w, r)
	case http.MethodPut, http.MethodPatch:
		UpdateUser(w, r)
	case http.MethodGet:
		ListUsers(w, r)
	case http.MethodDelete:
		DeleteUser(w, r)
	default:
		w.Write([]byte("oh no, unhandled endpoint found: " + r.Method))
	}
	w.Write([]byte("\n"))
}

func AddRoutes(router *http.ServeMux, ms Middleware) {
	router.HandleFunc("/v1/user/", RequestLogger(ms(endpoint)))
}
