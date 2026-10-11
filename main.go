package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", createTask)
	fmt.Println("Listening to server : 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Print(err)
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
