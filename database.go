package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./frappe.db")
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	err = createTables()
	if err != nil {
		return err
	}

	return nil
}

func createTables() error {
	createDoctypeTable := `
	CREATE TABLE IF NOT EXISTS doctypes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);`

	_, err := db.Exec(createDoctypeTable)
	if err != nil {
		return err
	}

	createFieldTable := `
	CREATE TABLE IF NOT EXISTS fields (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		doctype_id INTEGER,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		label TEXT NOT NULL,
		required BOOLEAN NOT NULL,
		FOREIGN KEY (doctype_id) REFERENCES doctypes(id)
	);`

	_, err = db.Exec(createFieldTable)
	if err != nil {
		return err
	}

	createPermissionTable := `
	CREATE TABLE IF NOT EXISTS permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		doctype_id INTEGER,
		permission TEXT NOT NULL,
		FOREIGN KEY (doctype_id) REFERENCES doctypes(id)
	);`

	_, err = db.Exec(createPermissionTable)
	if err != nil {
		return err
	}

	createDocumentTable := `
	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		doctype_id INTEGER,
		data TEXT NOT NULL,
		FOREIGN KEY (doctype_id) REFERENCES doctypes(id)
	);`

	_, err = db.Exec(createDocumentTable)
	if err != nil {
		return err
	}

	return nil
}
