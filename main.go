package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

var database *Database

func main() {
	database = startDb()
	defer database.db.Close()
	//createTable()

	startHttpServer()
}

func startDb() *Database {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	return &Database{db: db}
}

func startHttpServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", createTask)
	fmt.Println("Listening to server : 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Print(err)
	}
}

func createTable() {
	sqlStmt := `
	create table task (id integer not null primary key, task text, status text, createdAt datetime);
	`
	_, err := database.db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func insertTask(taskName string) {
	sqlStmt := `
	insert into task (task, status, createdAt) values (?, ?, ?);
	`
	_, err := database.db.Exec(sqlStmt, taskName, "CREATED", time.Now().UTC())
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON : "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("creating a new task with Taskname : ", req.TaskName)
	insertTask(req.TaskName)

	resp := CreateTaskResp{req.TaskName, "CREATED", time.Now().UTC()}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

type CreateTaskRequest struct {
	TaskName string `json:"taskName" binding:"required"`
}

type CreateTaskResp struct {
	TaskName  string    `json:"taskName" binding:"required"`
	Status    string    `json:"status" binding:"required"`
	CreatedAt time.Time `json:"createdAt" binding:"required"`
}
