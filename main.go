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
	createTableIfNeeded()

	startHttpServer()
}

func startDb() *Database {
	db, err := sql.Open("sqlite3", "/Users/gurusankar/Work/Codebase/go/workspace/todo/local-dev.db")
	if err != nil {
		log.Fatal(err)
	}
	return &Database{db: db}
}

func startHttpServer() {
	dir := http.Dir("./frontend")
	fs := http.FileServer(dir)
	mux := http.NewServeMux()

	mux.Handle("/task-manager/", http.StripPrefix("/task-manager/", fs))
	mux.HandleFunc("POST /api/tasks", createTaskHandler)
	mux.HandleFunc("GET /api/tasks", fetchTasksHandler)

	fmt.Println("Listening to server : 8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		fmt.Print(err)
	}
}

func createTableIfNeeded() {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS tasks (
		id integer not null primary key, 
		taskname text, 
		labels text,
		status text, 
		createdAt datetime);
	`
	_, err := database.db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func insertTask(taskName string, lables []string) (lastInsertedId int64) {
	sqlStmt := `
		insert into tasks (taskname, labels, status, createdAt) values (?, ?, ?, ?);
	`
	// Convert the slice to a JSON string
	lables_, err_ := json.Marshal(lables)
	if err_ != nil {
		log.Fatal(err_)
		return
	}

	fmt.Println(string(lables_))
	out, err := database.db.Exec(sqlStmt, taskName, string(lables_), "CREATED", time.Now().UTC())
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	lastInsertedId, err = out.LastInsertId()
	fmt.Println("Task saved successfully with Taskname : ", lastInsertedId)
	return lastInsertedId
}

func fetchTasks() []TaskResp {
	sqlStmt := `
		select id, taskname, labels, status, createdAt from tasks order by id desc;
	`
	rows, err := database.db.Query(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil
	}
	defer rows.Close()

	var tasks []TaskResp
	for rows.Next() {
		var task TaskResp
		var labelsRaw sql.NullString
		var timeVal time.Time
		err = rows.Scan(&task.Id, &task.TaskName, &labelsRaw, &task.Status, &timeVal)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return nil
		}
		if labelsRaw.Valid && labelsRaw.String != "" {
			err = json.Unmarshal([]byte(labelsRaw.String), &task.Labels)
			if err != nil {
				fmt.Print("failed to parse Labels JSON: %w", err)
				return nil
			}
		}
		task.CreatedAt = timeVal.Format("2006-01-02 15:04:05") // this formatting numbers needs to be exactly this.
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil
	}

	return tasks
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON : "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("creating a new task with Taskname : ", req.TaskName)
	lastInsertedId := insertTask(req.TaskName, req.Labels)
	resp := TaskResp{Id: lastInsertedId, TaskName: req.TaskName, Status: "CREATED", CreatedAt: time.Now().UTC().Format("YYYY-MM-dd"), Labels: req.Labels}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

func fetchTasksHandler(w http.ResponseWriter, r *http.Request) {
	output := fetchTasks()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(output)
	if err != nil {
		return
	}

}

type CreateTaskRequest struct {
	TaskName string   `json:"taskName" binding:"required"`
	Labels   []string `json:"labels" binding:"optional"`
}

type TaskResp struct {
	Id        int64    `json:"id" binding:"required"`
	TaskName  string   `json:"taskName" binding:"required"`
	Status    string   `json:"status" binding:"required"`
	Labels    []string `json:"labels" binding:"optional"`
	CreatedAt string   `json:"createdAt" binding:"required"`
}
