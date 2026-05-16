package api

import (
	"log"
	"net/http"

	"kingultron99.com/RoleCall/middleware"
)

type Route struct {
	Authenticated bool
	Scope         string
	Handler       func(w http.ResponseWriter, r *http.Request)
}

var (
	mux       *http.ServeMux
	mapRoutes = make(map[string]Route)
)

func StartApi() {
	mux = http.NewServeMux()
	registerApiRoutes()
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP server listening on :8080")

}

func registerApiRoutes() {
	for route, handler := range mapRoutes {
		if handler.Authenticated {
			mux.Handle(route, middleware.JWTAuth(handler.Scope, http.HandlerFunc(handler.Handler)))
		} else {
			mux.HandleFunc(route, handler.Handler)
		}
		log.Println("Registered api route: " + route)
	}
}
