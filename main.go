package main

import (
	"fmt"
	"kratos-rbac/constants"
	"kratos-rbac/routes"
	"net/http"
)

func runServer(port string, mux *http.ServeMux) {
	fmt.Println("running a server on", port)
	fmt.Println(http.ListenAndServe(":"+port, mux))
	fmt.Println("server exited")
}

func main() {
	mux := http.NewServeMux()
	routes.AddRoutes(mux, routes.AdminMiddleware)

	runServer(constants.DefaultPort, mux)
}
