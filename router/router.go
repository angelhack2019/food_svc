package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"GetFood",
		"GET",
		"/food/{uuid}",
		getFood,
	},
	Route{
		"GetFoods",
		"GET",
		"/foods",
		getFoods,
	},
	Route{
		"DeleteFood",
		"DELETE",
		"/food/{uuid}",
		deleteFood,
	},
	Route{
		"UpdateFood",
		"PUT",
		"/food/{uuid}",
		updateFood,
	},
	Route{
		"ShareFood",
		"POST",
		"/food",
		shareFood,
	},
}
