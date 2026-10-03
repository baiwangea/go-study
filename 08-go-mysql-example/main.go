package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// User represents the structure of a user in our database.
type User struct {
	ID    int
	Name  string
	Email string
	Age   int
}

func main() {
	// Data Source Name (DSN)
	// Format: user:password@tcp(host:port)/dbname
	// IMPORTANT: Replace with your actual MySQL connection details.
	dsn := "root:password@tcp(127.0.0.1:3306)/testdb?parseTime=true"

	// Open a connection to the database.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Set connection pool settings.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Ping the database to verify the connection is alive.
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v. Please check your DSN and ensure MySQL is running.", err)
	}
	fmt.Println("Successfully connected to MySQL!")

	// Create the users table if it doesn't exist.
	createTable(db)

	// Run CRUD examples.
	userID := createExample(db)
	readExample(db, userID)
	updateExample(db, userID)
	deleteExample(db, userID)
}

func createTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(50) NOT NULL,
		email VARCHAR(50) NOT NULL UNIQUE,
		age INT
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	fmt.Println("Table 'users' created or already exists.")
}

func createExample(db *sql.DB) int64 {
	fmt.Println("\n--- Create Example ---")

	// Use a prepared statement for insertion to prevent SQL injection.
	stmt, err := db.Prepare("INSERT INTO users(name, email, age) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Execute the prepared statement.
	result, err := stmt.Exec("Alice", "alice@example.com", 28)
	if err != nil {
		// If the email already exists, we might get an error. For this example, we'll ignore it.
		log.Printf("Could not insert new user (might already exist): %v", err)
		// Let's query the ID of the existing user instead.
		var id int64
		db.QueryRow("SELECT id FROM users WHERE email = ?", "alice@example.com").Scan(&id)
		return id
	}

	// Get the ID of the newly inserted user.
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted a new user with ID: %d\n", lastID)
	return lastID
}

func readExample(db *sql.DB, id int64) {
	fmt.Println("\n--- Read Example ---")

	// Query a single row.
	var user User
	// Use QueryRow for single-row queries.
	err := db.QueryRow("SELECT id, name, email, age FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name, &user.Email, &user.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with ID %d", id)
			return
		}
		log.Fatal(err)
	}
	fmt.Printf("Found a single user: %+v\n", user)

	// Query multiple rows.
	rows, err := db.Query("SELECT id, name FROM users LIMIT 2")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Found multiple users (first 2):")
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  - ID: %d, Name: %s\n", u.ID, u.Name)
	}
}

func updateExample(db *sql.DB, id int64) {
	fmt.Println("\n--- Update Example ---")

	// Use a prepared statement for updating.
	stmt, err := db.Prepare("UPDATE users SET age = ? WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(30, id)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated user with ID %d. Rows affected: %d\n", id, rowsAffected)
}

func deleteExample(db *sql.DB, id int64) {
	fmt.Println("\n--- Delete Example ---")

	// Use a prepared statement for deletion.
	stmt, err := db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted user with ID %d. Rows affected: %d\n", id, rowsAffected)
}
