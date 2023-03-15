package main

import (
	"kratos-rbac/constants"
	"kratos-rbac/routes"
	"net/http"
)

func runServer(port string, mux *http.ServeMux) {
	http.ListenAndServe(":"+port, mux)
}

func main() {
	mux := http.NewServeMux()
	routes.AddRoutes(mux, routes.AdminMiddleware)

	runServer(constants.DefaultPort, mux)
}
