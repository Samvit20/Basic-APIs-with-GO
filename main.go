package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"io"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/denisenkom/go-mssqldb"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	// Query the database
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Prepare a slice to hold the users
	var users []User

	// Process the query results
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Name, &user.Age)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

	// Convert users slice to JSON
	jsonUsers, err := json.Marshal(users)
	if err != nil {
		log.Fatal(err)
	}

	// Set response headers and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonUsers)
}

func getUsersById(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	id := params["id"]

	query := "SELECT * FROM users WHERE id = @id"

	// Prepare a slice to hold the user
	var user User

	err := db.QueryRow(query, sql.Named("id", id)).Scan(&user.Id, &user.Name, &user.Age)
	if err != nil {
		log.Fatal(err)
	}

	// Convert user slice to JSON
	jsonUser, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}

	// Set response headers and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonUser)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		log.Fatal(err)
	}
	query := "INSERT INTO users (name, age) VALUES (@Name, @Age)"

	_, err = db.Exec(query, sql.Named("Name", newUser.Name), sql.Named("Age", newUser.Age))
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User added successfully"))
}

func deleteUserById(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request URL parameters
	params := mux.Vars(r)
	id := params["id"]

	// Prepare the DELETE query
	query := "DELETE FROM users WHERE id = @id"

	// Execute the query to delete the user
	_, err := db.Exec(query, sql.Named("id", id))
	if err != nil {
		log.Fatal(err)
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted successfully"))
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // Limit the maximum upload size to 10MB
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Error parsing multipart form:", err)
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Error retrieving file from form:", err)
		return
	}
	defer file.Close()

	// Create a new file in the server's filesystem
	f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error creating file:", err)
		return
	}
	defer f.Close()

	// Copy the uploaded file's data to the newly created file
	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error copying file data:", err)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "File uploaded successfully")
}

var db *sql.DB

func main() {
	// Connection string
	connectionString := "sqlserver://sa:Samvit1234!@localhost:1433?database=master"

	// Open connection
	var err error
	db, err = sql.Open("sqlserver", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to the database!")

	// Create router and define routes
	router := mux.NewRouter()
	router.HandleFunc("/users", getUsers).Methods("GET")
	router.HandleFunc("/users/{id}", getUsersById).Methods("GET")
	router.HandleFunc("/addusers", addUser).Methods("POST")
	router.HandleFunc("/deleteuser/{id}", deleteUserById).Methods("DELETE")
	router.HandleFunc("/uploadfile", uploadFile).Methods("POST")
	// Serve static files (uploaded images) from the current directory
	fs := http.FileServer(http.Dir("."))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", router))
}