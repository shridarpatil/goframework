package main

import (
	"github.com/gorilla/mux"
)

func registerRoutes(r *mux.Router) {

	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")

	r.HandleFunc("/", authMiddleware(homeHandler)).Methods("GET")
	r.HandleFunc("/doctypes", authMiddleware(doctypeListHandler)).Methods("GET")
	r.HandleFunc("/doctype/new", authMiddleware(doctypeNewHandler)).Methods("GET", "POST")
	r.HandleFunc("/doctype/{name}", authMiddleware(doctypeHandler)).Methods("GET")
	r.HandleFunc("/doctype/{name}/edit", authMiddleware(doctypeEditHandler)).Methods("GET", "POST")
	r.HandleFunc("/doctype/{name}/documents", authMiddleware(documentListHandler)).Methods("GET")
	r.HandleFunc("/doctype/{name}/document/new", authMiddleware(documentNewHandler)).Methods("GET", "POST")
	r.HandleFunc("/doctype/{name}/document/{id}", authMiddleware(documentEditHandler)).Methods("GET", "POST")

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/documents", apiCreateDocument).Methods("POST")
	api.HandleFunc("/documents/{doctype}/{id}", apiGetDocument).Methods("GET")
	api.HandleFunc("/documents/{doctype}/{id}", apiUpdateDocument).Methods("PUT")
	api.HandleFunc("/documents/{doctype}/{id}", apiDeleteDocument).Methods("DELETE")
	api.HandleFunc("/documents/{doctype}", apiListDocuments).Methods("GET")
}
