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

	// Create Role doctype
	err = createRoleDoctype()
	if err != nil {
		return err
	}

	// Create default roles
	defaultRoles := []Document{
		{DoctypeName: "Role", Data: map[string]interface{}{"name": "Admin", "description": "Administrator role"}},
		{DoctypeName: "Role", Data: map[string]interface{}{"name": "User", "description": "Regular user role"}},
		{DoctypeName: "Role", Data: map[string]interface{}{"name": "Guest", "description": "Guest user role"}},
	}

	for _, role := range defaultRoles {
		err = createDocument(&role)
		if err != nil {
			return err
		}
	}

	// Create User doctype
	err = createUserDoctype()
	if err != nil {
		return err
	}

	// Check if users already exist
	users, err := getDocuments("User")
	if err != nil {
		return err
	}

	if len(users) == 0 {
		// Create Admin user
		adminUser := Document{
			DoctypeName: "User",
			Data: map[string]interface{}{
				"username": "admin",
				"password": "admin123", // In a real application, this should be hashed
				"is_admin": true,
				"role":     "Admin",
			},
		}
		err = createDocument(&adminUser)
		if err != nil {
			return err
		}

		// Create Guest user
		guestUser := Document{
			DoctypeName: "User",
			Data: map[string]interface{}{
				"username": "guest",
				"password": "guest123", // In a real application, this should be hashed
				"is_admin": false,
				"role":     "Guest",
			},
		}
		err = createDocument(&guestUser)
		if err != nil {
			return err
		}
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

	createFieldPermissionsTable := `
    CREATE TABLE IF NOT EXISTS field_permissions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        field_id INTEGER,
        permission TEXT NOT NULL,
        FOREIGN KEY (field_id) REFERENCES fields(id)
    );`

	_, err = db.Exec(createFieldPermissionsTable)
	if err != nil {
		return err
	}

	createDoctypePermissionsTable := `
    CREATE TABLE IF NOT EXISTS doctype_permissions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        doctype_id INTEGER,
        permission TEXT NOT NULL,
        FOREIGN KEY (doctype_id) REFERENCES doctypes(id)
    );`

	_, err = db.Exec(createDoctypePermissionsTable)
	if err != nil {
		return err
	}

	return nil
}
