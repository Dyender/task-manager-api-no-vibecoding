package main

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID          int       `json:"id"`
	UserID      int       `json:"userId"`
	Name        string    `json:"name"`
	Desc        string    `json:"desc"`
	IsCompleted bool      `json:"isCompleted"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TaskUpdate struct {
	Name        *string `json:"name"`
	Desc        *string `json:"desc"`
	IsCompleted *bool   `json:"isCompleted"`
}

type AllTasks struct {
	Tasks []Task `json:"tasks"`
}

type Db struct {
	dbConnect *pgxpool.Pool
}

func NewDb(db *pgxpool.Pool) *Db {
	return &Db{db}
}

func (db *Db) AddTaskDB(ctx context.Context, task Task) (Task, error) {
	err := db.dbConnect.QueryRow(
		ctx,
		`insert into tasks (taskName, userID, taskDesc, isCompleted) VALUES ($1, $2, $3, $4) returning id, createdAt, updatedAt`,
		task.Name,
		task.UserID,
		task.Desc,
		task.IsCompleted,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (db *Db) GetTaskByTaskIdDB(ctx context.Context, id, userID int) (Task, error) {
	var task Task
	err := db.dbConnect.QueryRow(
		ctx,
		`select * from tasks where id = $1 and userID = $2`,
		id,
		userID,
	).Scan(&task.ID, &task.UserID, &task.Name, &task.Desc, &task.IsCompleted, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (db *Db) GetTasksByUserIdDB(ctx context.Context, userId int) (AllTasks, error) {
	var tasks []Task
	rows, err := db.dbConnect.Query(
		ctx,
		`select * from tasks where userID = $1`,
		userId,
	)
	if err != nil {
		return AllTasks{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.UserID, &task.Name, &task.Desc, &task.IsCompleted, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return AllTasks{}, err
		}

		tasks = append(tasks, task)
	}
	return AllTasks{tasks}, nil
}

func (db *Db) GetAllTasksDB(ctx context.Context) ([]*Task, error) {
	tasks := []*Task{}

	rows, err := db.dbConnect.Query(
		ctx,
		`select * from tasks`)
	if err != nil {
		return []*Task{}, err
	}
	defer rows.Close()
	for rows.Next() {
		task := Task{}
		err = rows.Scan(&task.ID, &task.Name, &task.Desc, &task.IsCompleted, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return []*Task{}, err
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}

func (db *Db) DeleteTaskById(ctx context.Context, id, userId int) (Task, error) {
	var task Task

	err := db.dbConnect.QueryRow(
		ctx,
		`delete from tasks where id = $1 and userId = $2 returning id, taskName, taskDesc, isCompleted ,createdAt, updatedAt`,
		id,
		userId,
	).Scan(&task.ID, &task.Name, &task.Desc, &task.IsCompleted, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (db *Db) UpdateTaskByIdDB(ctx context.Context, task TaskUpdate, id, userId int) error {
	query := sq.Update("tasks").PlaceholderFormat(sq.Dollar)
	if task.Name != nil && *task.Name != "" {
		query = query.Set("taskName", *task.Name)
	}
	if task.Desc != nil {
		query = query.Set("taskDesc", *task.Desc)
	}
	if task.IsCompleted != nil {
		query = query.Set("isCompleted", *task.IsCompleted)
	}
	query = query.Set("updatedAt", time.Now())
	query = query.Where(sq.Eq{"id": id})
	query = query.Where(sq.Eq{"userID": userId})

	querySql, args, err := query.ToSql()

	if err != nil {
		return err
	}

	_, err = db.dbConnect.Query(
		ctx,
		querySql,
		args...,
	)

	return err
}
