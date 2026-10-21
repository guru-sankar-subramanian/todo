package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/guru-sankar-subramanian/todo/internal/db"
	"github.com/guru-sankar-subramanian/todo/internal/models"
)

func Init(prodProfileEnabled bool, embeddedUI embed.FS) http.Handler {
	mux := http.NewServeMux()

	// UI
	var uiFS http.FileSystem
	if !prodProfileEnabled {
		uiFS = http.FS(os.DirFS("./frontend")) // live reload from disk while developing
	} else {
		sub, err := fs.Sub(embeddedUI, "frontend") // strip the "frontend/" prefix
		if err != nil {
			log.Fatal(err)
		}
		uiFS = http.FS(sub)
	}
	mux.Handle("/", http.FileServer(uiFS))

	// API
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/tasks", fetchTasksHandler)
	mux.HandleFunc("POST /api/tasks", createTaskHandler)
	mux.HandleFunc("PUT /api/tasks", updateTasksHandler)

	return mux
}


func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON : "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("creating a new task with Taskname : ", req.TaskName)
	lastInsertedId := db.InsertTask(req.TaskName, req.Labels)
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
	output := db.FetchTasks()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(output)
	if err != nil {
		return
	}
}

// status = TODO, IN_PROGRESS, DONE
func updateTasksHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateTaskRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON : "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("updating existing task with Taskname : ", req.TaskName)
	lastInsertedId := db.UpdateTask(req)
	resp := fmt.Sprintf("updated task with ID : %d and total rows impacted : %d", req.Id, lastInsertedId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}
