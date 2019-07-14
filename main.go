package main

import (
	"github.com/angelhack2019/food_svc/router"
	"log"
	"net/http"
)

func main() {

	router := router.NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
