package main

import (
	"github.com/gorilla/mux"
)

func registerRoutes(r *mux.Router) {
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/doctypes", doctypeListHandler).Methods("GET")
	r.HandleFunc("/doctype/new", doctypeNewHandler).Methods("GET", "POST")
	r.HandleFunc("/doctype/{name}", doctypeHandler).Methods("GET")
	r.HandleFunc("/doctype/{name}/edit", doctypeEditHandler).Methods("GET", "POST")
	r.HandleFunc("/doctype/{name}/documents", documentListHandler).Methods("GET")
	r.HandleFunc("/doctype/{name}/document/new", documentNewHandler).Methods("GET", "POST")
	r.HandleFunc("/doctype/{name}/document/{id}", documentEditHandler).Methods("GET", "POST")

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/documents", apiCreateDocument).Methods("POST")
	api.HandleFunc("/documents/{doctype}/{id}", apiGetDocument).Methods("GET")
	api.HandleFunc("/documents/{doctype}/{id}", apiUpdateDocument).Methods("PUT")
	api.HandleFunc("/documents/{doctype}/{id}", apiDeleteDocument).Methods("DELETE")
	api.HandleFunc("/documents/{doctype}", apiListDocuments).Methods("GET")
}
