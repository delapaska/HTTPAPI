package main

import (
	"fmt"
	"net/http"

	"github.com/delapaska/AvitoTest/db"
	"github.com/delapaska/AvitoTest/db/segments"
	"github.com/delapaska/AvitoTest/handlers"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	conDB := db.DBConnect()
	repo := segments.New(conDB, &segments.Repository{})
	go repo.CheckTTL()

	http.HandleFunc("/download/", handlers.DownloadCSVHandler)
	r.HandleFunc("/segment/add", handlers.CreateHandler(repo)).Methods("POST")
	r.HandleFunc("/segment/delete", handlers.DeleteHandler(repo)).Methods("DELETE")
	r.HandleFunc("/user/add", handlers.AddUserHandler(repo)).Methods("POST")
	r.HandleFunc("/user/delete", handlers.DeleteUserHandler(repo)).Methods("POST")
	r.HandleFunc("/user/return", handlers.ReturnSegmentHandler(repo)).Methods("GET")
	r.HandleFunc("/user/distribute", handlers.DistributeUsersHandler(repo)).Methods("POST")
	r.HandleFunc("/user/history", handlers.GetUserHistoryHandler(repo)).Methods("GET")
	http.Handle("/", r)
	port := ":8080"
	http.ListenAndServe(port, nil)
	fmt.Printf("Сервер слушает на порту %s...\n", port)
	 
	defer conDB.Close()
}
