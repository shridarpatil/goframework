package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

var templates map[string]*template.Template

var templateFuncs = template.FuncMap{
	"contains": func(slice []string, item string) bool {
		for _, s := range slice {
			if s == item {
				return true
			}
		}
		return false
	},
}

func init() {
	templates = make(map[string]*template.Template)
	templatesDir := "templates/"
	layouts, err := filepath.Glob(templatesDir + "*.html")
	if err != nil {
		panic(err)
	}
	for _, layout := range layouts {
		files := []string{layout, templatesDir + "base.html"}
		templates[filepath.Base(layout)] = template.Must(template.New(filepath.Base(layout)).Funcs(templateFuncs).ParseFiles(files...))
	}
}

// PageData is a struct to hold common page data
type PageData struct {
	Title   string
	Content interface{}
	User    *Document
}

type DocumentFormData struct {
	Doctype  Doctype
	Document *Document // Note the pointer here
	IsNew    bool
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := getUserByUsername(username)
		if err != nil {
			renderTemplate(w, r, "login.html", map[string]interface{}{"Error": "Invalid username or password"})
			return
		}

		storedPassword, ok := user.Data["password"].(string)
		if !ok {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		if err != nil {
			renderTemplate(w, r, "login.html", map[string]interface{}{"Error": "Invalid username or password"})
			return
		}

		// Set user as authenticated
		session.Values["authenticated"] = true
		session.Values["user_id"] = user.ID
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "login.html", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Values["user_id"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Home",
		Content: "Welcome to the Frappe Framework",
	}
	renderTemplate(w, r, "home.html", data)
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
	renderTemplate(w, r, "doctype_list.html", data)
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
		Title:   "New Doctype",
		Content: nil,
	}
	renderTemplate(w, r, "doctype_new.html", data)

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
	renderTemplate(w, r, "doctype.html", data)
}

func doctypeEditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	doctype, err := getDoctypeByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roles, err := getRoles()
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

		log.Printf("Received form data: %+v", r.Form)

		doctype.Name = r.FormValue("name")
		doctype.Fields = []Field{}
		doctype.Permissions = r.Form["permissions"]

		fieldIDs := r.Form["field_id"]
		fieldNames := r.Form["field_name"]
		fieldTypes := r.Form["field_type"]
		fieldLabels := r.Form["field_label"]
		fieldRequired := r.Form["field_required"]
		fieldPermissions := r.Form["field_permissions"]

		// Find the minimum length of all field-related slices
		minLen := len(fieldNames)
		if len(fieldTypes) < minLen {
			minLen = len(fieldTypes)
		}
		if len(fieldLabels) < minLen {
			minLen = len(fieldLabels)
		}
		if len(fieldIDs) < minLen {
			minLen = len(fieldIDs)
		}

		for i := 0; i < minLen; i++ {
			id, _ := strconv.ParseInt(fieldIDs[i], 10, 64)
			required := false
			for _, req := range fieldRequired {
				if req == fieldNames[i] {
					required = true
					break
				}
			}
			field := Field{
				ID:          id,
				DoctypeID:   doctype.ID,
				Name:        fieldNames[i],
				Type:        fieldTypes[i],
				Label:       fieldLabels[i],
				Required:    required,
				Permissions: []string{},
			}
			if i < len(fieldPermissions) {
				field.Permissions = strings.Fields(fieldPermissions[i])
			}
			doctype.Fields = append(doctype.Fields, field)
		}

		doctypeJSON, _ := json.MarshalIndent(doctype, "", "  ")
		log.Printf("Doctype to be updated: %s", string(doctypeJSON))

		err = updateDoctype(&doctype)
		if err != nil {
			log.Printf("Error updating doctype: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Doctype updated successfully")

		http.Redirect(w, r, "/doctypes", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title: "Edit " + doctype.Name,
		Content: struct {
			Doctype Doctype
			Roles   []string
		}{
			Doctype: doctype,
			Roles:   roles,
		},
	}

	renderTemplate(w, r, "doctype_edit.html", data)
}

func getRoles() ([]string, error) {
	// Implement this function to fetch roles from your database
	// For now, we'll return a static list
	return []string{"Admin", "User", "Guest"}, nil
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

	renderTemplate(w, r, "document_list.html", data)
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

func renderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	t, ok := templates[tmpl]
	if !ok {
		http.Error(w, fmt.Sprintf("Template %s not found", tmpl), http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["user_id"].(int)
	var user *Document
	if ok && userID != 0 {
		user, _ = getUserByID(userID)
	}

	var dataMap map[string]interface{}

	switch v := data.(type) {
	case PageData:
		dataMap = map[string]interface{}{
			"Title":   v.Title,
			"Content": v.Content,
		}
	case map[string]interface{}:
		dataMap = v
	default:
		dataMap = make(map[string]interface{})
	}

	dataMap["User"] = user

	buf := &bytes.Buffer{}
	err := t.ExecuteTemplate(buf, "base", dataMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
