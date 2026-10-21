package models


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

type UpdateTaskRequest struct {
	Id       int64    `json:"id" binding:"required"`
	TaskName string   `json:"taskName" binding:"required"`
	Labels   []string `json:"labels" binding:"optional"`
	Status   string   `json:"status" binding:"required"`
}
