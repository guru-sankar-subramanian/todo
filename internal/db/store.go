package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"path/filepath"
	"time"

	"github.com/guru-sankar-subramanian/todo/internal/models"
	_ "modernc.org/sqlite"
)

type Database struct {
	Db *sql.DB
}

var database *Database

func StartDb(dataDir string) *Database {
	dbPath := filepath.Join(dataDir, "data.db")
	log.Println("DB is located in ", dbPath)
	db, err := sql.Open("sqlite", dbPath)
	// db.SetMaxOpenConns(1)
	// _, _ = db.Exec("PRAGMA journal_mode=WAL;")
	// _, _ = db.Exec("PRAGMA busy_timeout=5000;")
	if err != nil {
		log.Fatal(err)
	}
	database = &Database{Db: db}
	createTableIfNeeded()
	return database
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
	_, err := database.Db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	log.Println("Tasks table created/ verified")
}

func UpdateTask(req models.UpdateTaskRequest) (lastInsertedId int64) {
	sqlStmt := `
		update tasks 
		set labels = ?, status = ?, updatedAt = ?, taskName = ?
		where id = ?;
	`
	// Convert the slice to a JSON string
	labels_, err_ := json.Marshal(req.Labels)
	if err_ != nil {
		log.Fatal(err_)
		return
	}

	log.Println("labels are :", string(labels_))
	out, err := database.Db.Exec(sqlStmt, string(labels_), req.Status, time.Now().UTC(), req.TaskName, req.Id)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	rowsAffected, err := out.RowsAffected()
	log.Println("Updated successfully with total affected rows ", rowsAffected)
	return rowsAffected
}

func InsertTask(taskName string, labels []string) (lastInsertedId int64) {
	sqlStmt := `
		insert into tasks (taskname, labels, status, createdAt) values (?, ?, ?, ?);
	`
	// Convert the slice to a JSON string
	labels_, err_ := json.Marshal(labels)
	if err_ != nil {
		log.Fatal(err_)
		return
	}

	log.Println(string(labels_))
	out, err := database.Db.Exec(sqlStmt, taskName, string(labels_), "TO_DO", time.Now().UTC())
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	lastInsertedId, err = out.LastInsertId()
	log.Println("Task saved successfully with lastInsertedId : ", lastInsertedId)
	return lastInsertedId
}

func FetchTasks() []models.TaskResp {
	sqlStmt := `
		select id, taskname, labels, status, createdAt from tasks order by id desc;
	`
	rows, err := database.Db.Query(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil
	}
	defer rows.Close()

	var tasks []models.TaskResp
	for rows.Next() {
		var task models.TaskResp
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
				log.Print("failed to parse Labels JSON: %w", err)
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
	log.Println("get tasks called and returned total rows ", len(tasks))
	return tasks
}
