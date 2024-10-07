package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	// "strconv"
	"strings"

	"github.com/gorilla/mux"
)

var templates map[string]*template.Template

func init() {
	templates = make(map[string]*template.Template)
	templatesDir := "templates/"
	layouts, err := filepath.Glob(templatesDir + "*.html")
	if err != nil {
		panic(err)
	}
	for _, layout := range layouts {
		templates[filepath.Base(layout)] = template.Must(template.ParseFiles(layout, templatesDir+"base.html"))
	}
}

// PageData is a struct to hold common page data
type PageData struct {
	Title   string
	Content interface{}
}

type DocumentFormData struct {
	Doctype  Doctype
	Document *Document // Note the pointer here
	IsNew    bool
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Home",
		Content: "Welcome to Frappe Framework",
	}
	renderTemplate(w, "home.html", data)
}

func doctypeListHandler(w http.ResponseWriter, r *http.Request) {
	doctypes, err := getDoctypes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "Doctypes",
		Content: struct {
			Doctypes []Doctype
		}{
			Doctypes: doctypes,
		},
	}
	renderTemplate(w, "doctype_list.html", data)
}

func doctypeNewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		newDoctype := Doctype{
			Name:        r.FormValue("name"),
			Permissions: r.Form["permissions"],
		}

		fieldNames := r.Form["field_name"]
		fieldTypes := r.Form["field_type"]
		fieldLabels := r.Form["field_label"]
		fieldRequired := r.Form["field_required"]

		for i := range fieldNames {
			field := Field{
				Name:     fieldNames[i],
				Type:     fieldTypes[i],
				Label:    fieldLabels[i],
				Required: len(fieldRequired) > i && fieldRequired[i] == "on",
			}
			newDoctype.Fields = append(newDoctype.Fields, field)
		}

		err = createDoctype(&newDoctype)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/doctypes", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title:   "Create New Doctype",
		Content: nil,
	}
	renderTemplate(w, "doctype_new.html", data)
}

func doctypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	doctype, err := getDoctypeByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: doctype.Name,
		Content: struct {
			Doctype Doctype
		}{
			Doctype: doctype,
		},
	}

	renderTemplate(w, "doctype.html", data)
}

func doctypeEditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	doctype, err := getDoctypeByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Update doctype name
		doctype.Name = r.FormValue("name")

		// Update fields
		doctype.Fields = []Field{}
		fieldNames := r.Form["field_name"]
		fieldTypes := r.Form["field_type"]
		fieldLabels := r.Form["field_label"]
		fieldRequired := r.Form["field_required"]
		fieldPermissions := r.Form["field_permissions"]

		for i := range fieldNames {
			field := Field{
				Name:     fieldNames[i],
				Type:     fieldTypes[i],
				Label:    fieldLabels[i],
				Required: contains(fieldRequired, fieldNames[i]),
			}
			if i < len(fieldPermissions) {
				field.Permissions = strings.Fields(fieldPermissions[i])
			}
			doctype.Fields = append(doctype.Fields, field)
		}

		// Update permissions
		doctype.Permissions = r.Form["permissions"]

		// Update the doctype in the database
		err = updateDoctype(&doctype)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/doctypes", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title: "Edit " + doctype.Name,
		Content: struct {
			Doctype Doctype
		}{
			Doctype: doctype,
		},
	}

	renderTemplate(w, "doctype_edit.html", data)
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func documentListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	documents, err := getDocuments(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: name + " Documents",
		Content: struct {
			DoctypeName string
			Documents   []Document
		}{
			DoctypeName: name,
			Documents:   documents,
		},
	}

	renderTemplate(w, "document_list.html", data)
}

func documentNewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	doctype, err := getDoctypeByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	doc := Document{
		DoctypeName: name,
		Data:        make(map[string]interface{}),
	}
	for _, field := range doctype.Fields {
		doc.Data[field.Name] = ""
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		for _, field := range doctype.Fields {
			doc.Data[field.Name] = r.FormValue(field.Name)
		}

		err = createDocument(&doc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/doctype/"+name+"/documents", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title: "New " + name + " Document",
		Content: DocumentFormData{
			Doctype:  doctype,
			Document: &doc,
			IsNew:    true,
		},
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/document_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func documentEditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	id := vars["id"]

	doctype, err := getDoctypeByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isNew := id == "new"
	var doc Document

	if !isNew {
		doc, err = getDocumentByID(name, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		doc = Document{
			DoctypeName: name,
			Data:        make(map[string]interface{}),
		}
		for _, field := range doctype.Fields {
			doc.Data[field.Name] = ""
		}
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		for _, field := range doctype.Fields {
			doc.Data[field.Name] = r.FormValue(field.Name)
		}

		if isNew {
			err = createDocument(&doc)
		} else {
			err = updateDocument(&doc)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/doctype/"+name+"/documents", http.StatusSeeOther)
		return
	}

	formData := DocumentFormData{
		Doctype:  doctype,
		Document: &doc,
		IsNew:    isNew,
	}

	data := PageData{
		Title:   fmt.Sprintf("%s %s Document", map[bool]string{true: "New", false: "Edit"}[isNew], name),
		Content: formData,
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/document_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, ok := templates[tmpl]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist.", tmpl), http.StatusInternalServerError)
		return
	}
	err := t.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
