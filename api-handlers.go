package main

import (
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
    "strconv"
)

func apiCreateDocument(w http.ResponseWriter, r *http.Request) {
    var doc Document
    err := json.NewDecoder(r.Body).Decode(&doc)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err = createDocument(&doc)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(doc)
}

func apiGetDocument(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    doctype := vars["doctype"]
    id := vars["id"]

    doc, err := getDocumentByID(doctype, id)
    if err != nil {
        http.Error(w, "Document not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(doc)
}

func apiUpdateDocument(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    doctype := vars["doctype"]
    idStr := vars["id"]

    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var updatedDoc Document
    err = json.NewDecoder(r.Body).Decode(&updatedDoc)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    updatedDoc.ID = id
    updatedDoc.DoctypeName = doctype

    err = updateDocument(&updatedDoc)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(updatedDoc)
}

func apiDeleteDocument(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    doctype := vars["doctype"]
    id := vars["id"]

    err := deleteDocument(doctype, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func apiListDocuments(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    doctype := vars["doctype"]

    docs, err := getDocuments(doctype)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(docs)
}
