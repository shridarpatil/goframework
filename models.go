package main

import (
	"database/sql"
	// "encoding/json"
	"fmt"
	"strings"
)

// Remove the db variable declaration from here

type Doctype struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Fields      []Field  `json:"fields"`
	Permissions []string `json:"permissions"`
}

type Field struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	Required    bool     `json:"required"`
	Permissions []string `json:"permissions"`
}

type Document struct {
	ID          int                    `json:"id"`
	DoctypeName string                 `json:"doctype_name"`
	Data        map[string]interface{} `json:"data"`
}

func getDoctypes() ([]Doctype, error) {
	rows, err := db.Query("SELECT id, name FROM doctypes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctypes []Doctype
	for rows.Next() {
		var dt Doctype
		err := rows.Scan(&dt.ID, &dt.Name)
		if err != nil {
			return nil, err
		}

		dt.Fields, err = getFields(dt.ID)
		if err != nil {
			return nil, err
		}

		dt.Permissions, err = getPermissions(dt.ID)
		if err != nil {
			return nil, err
		}

		doctypes = append(doctypes, dt)
	}

	return doctypes, nil
}

func getFields(doctypeID int) ([]Field, error) {
	rows, err := db.Query("SELECT id, name, type, label, required FROM fields WHERE doctype_id = ?", doctypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []Field
	for rows.Next() {
		var f Field
		err := rows.Scan(&f.ID, &f.Name, &f.Type, &f.Label, &f.Required)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}

	return fields, nil
}

func getPermissions(doctypeID int) ([]string, error) {
	rows, err := db.Query("SELECT permission FROM permissions WHERE doctype_id = ?", doctypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		err := rows.Scan(&p)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	return permissions, nil
}

func createDoctype(dt *Doctype) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert into doctypes table
	result, err := tx.Exec("INSERT INTO doctypes (name) VALUES (?)", dt.Name)
	if err != nil {
		return err
	}

	doctypeID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Insert fields
	for _, field := range dt.Fields {
		_, err = tx.Exec("INSERT INTO fields (doctype_id, name, type, label, required) VALUES (?, ?, ?, ?, ?)",
			doctypeID, field.Name, field.Type, field.Label, field.Required)
		if err != nil {
			return err
		}
	}

	// Insert permissions
	for _, permission := range dt.Permissions {
		_, err = tx.Exec("INSERT INTO permissions (doctype_id, permission) VALUES (?, ?)",
			doctypeID, permission)
		if err != nil {
			return err
		}
	}

	// Create a new table for the doctype
	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (id INTEGER PRIMARY KEY AUTOINCREMENT", dt.Name)
	for _, field := range dt.Fields {
		sqlType := getSQLType(field.Type)
		createTableQuery += fmt.Sprintf(", `%s` %s", field.Name, sqlType)
	}
	createTableQuery += ")"

	_, err = tx.Exec(createTableQuery)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func getSQLType(fieldType string) string {
	switch fieldType {
	case "string":
		return "TEXT"
	case "text":
		return "TEXT"
	case "integer":
		return "INTEGER"
	case "float":
		return "REAL"
	case "boolean":
		return "INTEGER"
	case "date":
		return "TEXT"
	case "datetime":
		return "TEXT"
	case "select":
		return "TEXT"
	default:
		return "TEXT"
	}
}

func getDocuments(doctypeName string) ([]Document, error) {
	doctype, err := getDoctypeByName(doctypeName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, %s FROM `%s`",
		strings.Join(getFieldNames(doctype.Fields), ", "),
		doctypeName)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		doc.DoctypeName = doctypeName
		doc.Data = make(map[string]interface{})

		// Create a slice to hold the values
		values := make([]interface{}, len(doctype.Fields)+1)
		values[0] = &doc.ID
		for i := range doctype.Fields {
			values[i+1] = new(interface{})
		}

		err := rows.Scan(values...)
		if err != nil {
			return nil, err
		}

		// Populate the doc.Data map
		for i, field := range doctype.Fields {
			doc.Data[field.Name] = *(values[i+1].(*interface{}))
		}

		documents = append(documents, doc)
	}

	return documents, nil
}

func createDocument(doc *Document) error {
	doctype, err := getDoctypeByName(doc.DoctypeName)
	if err != nil {
		return err
	}

	columns := []string{}
	values := []interface{}{}
	placeholders := []string{}

	for _, field := range doctype.Fields {
		if value, ok := doc.Data[field.Name]; ok {
			columns = append(columns, field.Name)
			values = append(values, value)
			placeholders = append(placeholders, "?")
		}
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		doc.DoctypeName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := db.Exec(query, values...)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	doc.ID = int(id)
	return nil
}

func updateDocument(doc *Document) error {
	doctype, err := getDoctypeByName(doc.DoctypeName)
	if err != nil {
		return err
	}

	updates := []string{}
	values := []interface{}{}

	for _, field := range doctype.Fields {
		if value, ok := doc.Data[field.Name]; ok {
			updates = append(updates, fmt.Sprintf("`%s` = ?", field.Name))
			values = append(values, value)
		}
	}

	values = append(values, doc.ID)

	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE id = ?",
		doc.DoctypeName,
		strings.Join(updates, ", "))

	_, err = db.Exec(query, values...)
	return err
}

func getDoctypeByName(name string) (Doctype, error) {
	var dt Doctype
	err := db.QueryRow("SELECT id, name FROM doctypes WHERE name = ?", name).Scan(&dt.ID, &dt.Name)
	if err != nil {
		return dt, err
	}

	dt.Fields, err = getFields(dt.ID)
	if err != nil {
		return dt, err
	}

	dt.Permissions, err = getPermissions(dt.ID)
	if err != nil {
		return dt, err
	}

	return dt, nil
}

func getDocumentByID(doctypeName, id string) (Document, error) {
	// First, get the doctype to know the fields
	doctype, err := getDoctypeByName(doctypeName)
	if err != nil {
		return Document{}, fmt.Errorf("error getting doctype: %v", err)
	}

	// Construct the query
	fieldNames := getFieldNames(doctype.Fields)
	query := fmt.Sprintf("SELECT id, %s FROM `%s` WHERE id = ?",
		strings.Join(fieldNames, ", "),
		doctypeName)

	// Execute the query
	var doc Document
	doc.DoctypeName = doctypeName
	doc.Data = make(map[string]interface{})

	var scanValues []interface{}
	scanValues = append(scanValues, &doc.ID)
	for range fieldNames {
		scanValues = append(scanValues, new(interface{}))
	}

	err = db.QueryRow(query, id).Scan(scanValues...)
	if err != nil {
		if err == sql.ErrNoRows {
			return Document{}, fmt.Errorf("document not found")
		}
		return Document{}, fmt.Errorf("error querying document: %v", err)
	}

	// Populate the doc.Data map
	for i, field := range doctype.Fields {
		doc.Data[field.Name] = *(scanValues[i+1].(*interface{}))
	}

	return doc, nil
}

func updateDoctype(dt *Doctype) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the old doctype
	oldDoctype, err := getDoctypeByID(dt.ID)
	if err != nil {
		return err
	}

	// Update doctype name
	if oldDoctype.Name != dt.Name {
		_, err = tx.Exec("UPDATE doctypes SET name = ? WHERE id = ?", dt.Name, dt.ID)
		if err != nil {
			return err
		}

		// Rename the table
		_, err = tx.Exec(fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`", oldDoctype.Name, dt.Name))
		if err != nil {
			return err
		}
	}

	// Update fields
	_, err = tx.Exec("DELETE FROM fields WHERE doctype_id = ?", dt.ID)
	if err != nil {
		return err
	}

	for _, field := range dt.Fields {
		_, err = tx.Exec("INSERT INTO fields (doctype_id, name, type, label, required) VALUES (?, ?, ?, ?, ?)",
			dt.ID, field.Name, field.Type, field.Label, field.Required)
		if err != nil {
			return err
		}
	}

	// Update table structure
	for _, newField := range dt.Fields {
		oldField := getFieldByName(oldDoctype.Fields, newField.Name)
		if oldField == nil {
			// Add new column
			sqlType := getSQLType(newField.Type)
			_, err = tx.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", dt.Name, newField.Name, sqlType))
			if err != nil {
				return err
			}
		} else if oldField.Type != newField.Type {
			// Change column type
			sqlType := getSQLType(newField.Type)
			_, err = tx.Exec(fmt.Sprintf("ALTER TABLE `%s` ALTER COLUMN `%s` %s", dt.Name, newField.Name, sqlType))
			if err != nil {
				return err
			}
		}
	}

	// Remove deleted fields
	for _, oldField := range oldDoctype.Fields {
		if getFieldByName(dt.Fields, oldField.Name) == nil {
			_, err = tx.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", dt.Name, oldField.Name))
			if err != nil {
				return err
			}
		}
	}

	// Update permissions
	_, err = tx.Exec("DELETE FROM permissions WHERE doctype_id = ?", dt.ID)
	if err != nil {
		return err
	}

	for _, permission := range dt.Permissions {
		_, err = tx.Exec("INSERT INTO permissions (doctype_id, permission) VALUES (?, ?)",
			dt.ID, permission)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getFieldByName(fields []Field, name string) *Field {
	for _, field := range fields {
		if field.Name == name {
			return &field
		}
	}
	return nil
}

func getDoctypeByID(id int) (Doctype, error) {
	var dt Doctype
	err := db.QueryRow("SELECT id, name FROM doctypes WHERE id = ?", id).Scan(&dt.ID, &dt.Name)
	if err != nil {
		return dt, err
	}

	dt.Fields, err = getFields(dt.ID)
	if err != nil {
		return dt, err
	}

	dt.Permissions, err = getPermissions(dt.ID)
	if err != nil {
		return dt, err
	}

	return dt, nil
}

func getFieldNames(fields []Field) []string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = field.Name
	}
	return names
}

func deleteDocument(doctypeName, id string) error {
	query := fmt.Sprintf("DELETE FROM `%s` WHERE id = ?", doctypeName)
	_, err := db.Exec(query, id)
	return err
}
