package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	baseDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed to locate macOS Application Support directory: %v", err)
	}
	appDataDir := filepath.Join(baseDir, "todo-manager")
	if err := os.MkdirAll(appDataDir, 0700); err != nil { // 0700 restricts access strictly to user
		log.Fatalf("Failed to create app data directory: %v", err)
	}

	dbPath := filepath.Join(appDataDir, "data.db")
	fmt.Println("DB is located in ", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return &Database{db: db}
}

func startHttpServer() {
	dir := http.Dir("./frontend")
	fs := http.FileServer(dir)
	mux := http.NewServeMux()

	mux.Handle("/todo-manager/", http.StripPrefix("/todo-manager/", fs))
	mux.HandleFunc("GET /api/tasks", fetchTasksHandler)
	mux.HandleFunc("POST /api/tasks", createTaskHandler)
	mux.HandleFunc("PUT /api/tasks", updateTasksHandler)

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
		createdAt datetime,
		updatedAt datetime);
	`
	_, err := database.db.Exec(sqlStmt)
	fmt.Println("tasks table created")
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func updateTask(id int64, status string, labels []string) (lastInsertedId int64) {
	sqlStmt := `
		update tasks 
		set labels = ?, status = ?, updatedAt = ?
		where id = ?;
	`
	// Convert the slice to a JSON string
	labels_, err_ := json.Marshal(labels)
	if err_ != nil {
		log.Fatal(err_)
		return
	}

	fmt.Println("labels are :", string(labels_))
	out, err := database.db.Exec(sqlStmt, string(labels_), status, time.Now().UTC(), id)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	rowsAffected, err := out.RowsAffected()
	fmt.Println("Updated successfully with total affected rows ", rowsAffected)
	return rowsAffected
}

func insertTask(taskName string, labels []string) (lastInsertedId int64) {
	sqlStmt := `
		insert into tasks (taskname, labels, status, createdAt) values (?, ?, ?, ?);
	`
	// Convert the slice to a JSON string
	labels_, err_ := json.Marshal(labels)
	if err_ != nil {
		log.Fatal(err_)
		return
	}

	fmt.Println(string(labels_))
	out, err := database.db.Exec(sqlStmt, taskName, string(labels_), "TODO", time.Now().UTC())
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
	// resp := TaskResp{Id: lastInsertedId, TaskName: req.TaskName, Status: "CREATED", CreatedAt: time.Now().UTC().Format("YYYY-MM-dd"), Labels: req.Labels}
	resp := fmt.Sprintf("created task with ID : %d", lastInsertedId)
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

// status = TODO, IN_PROGRESS, DONE
func updateTasksHandler(w http.ResponseWriter, r *http.Request) {
	var req updateTaskRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON : "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("creating a new task with Taskname : ", req.Status)
	lastInsertedId := updateTask(req.Id, req.Status, req.Labels)
	resp := fmt.Sprintf("updated task with ID : %d and total rows impacted : %d", req.Id, lastInsertedId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
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

type updateTaskRequest struct {
	Id       int64    `json:"id" binding:"required"`
	TaskName string   `json:"taskName" binding:"required"`
	Labels   []string `json:"labels" binding:"optional"`
	Status   string   `json:"status" binding:"required"`
}
