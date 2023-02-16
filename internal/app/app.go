package app

import (
	"log"
	"net/http"
	"netmts/internal/controller"

	"github.com/gorilla/mux"
)

func Run() {
	router := mux.NewRouter()
	router.HandleFunc("/api", controller.HandleConnection)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	http.Handle("/", router)
	/*
		fileServer := http.FileServer(http.Dir("./web/"))
		router.Handle("/", http.NotFoundHandler())
		router.Handle("/", http.StripPrefix("/", fileServer))
	*/
	server := &http.Server{
		Addr:    ":8282",
		Handler: router,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
