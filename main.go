package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {

	dbCon, err := OpenDb()

	if err != nil {
		log.Fatal(err)
	}

	db := Db{dbCon}

	http.HandleFunc("/checkStatus/", func(w http.ResponseWriter, r *http.Request) {
		message := r.URL.Query().Get("msg")
		if message == "" {
			fmt.Fprint(w, "server alive")
			return
		}
		fmt.Fprintf(w, "Echo %s", message)
	})

	http.HandleFunc("POST /task", db.Add)

	http.HandleFunc("GET /task/", db.GetTask)

	http.HandleFunc("GET /tasks", db.GetTasks)

	http.HandleFunc("DELETE /task/", db.Delete)

	http.HandleFunc("PUT /task/", db.Update)

	http.HandleFunc("POST /user/reg", db.Register)

	http.HandleFunc("POST /user/login", db.Login)

	http.HandleFunc("DELETE /user", db.DeleteUser)

	http.HandleFunc("PUT /user", db.UpdateUserById)

	http.ListenAndServe(":8080", nil)
}

func AuthUser(w http.ResponseWriter, r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, errors.New("unauthorized error")
	}
	tokenString := authHeader[7:]

	mySecretToken := os.Getenv("MY_SECRET_TOKEN")
	if mySecretToken == "" {
		mySecretToken = "secret"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(mySecretToken), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token error")
	}
	claims := token.Claims.(jwt.MapClaims)

	userIdFloat := claims["user_id"].(float64)

	userId := int(userIdFloat)

	return userId, nil

}

func (db *Db) Add(w http.ResponseWriter, r *http.Request) {
	userId, err := AuthUser(w, r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	body := r.Body
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "body read error")
		return
	}

	taskToCreate := Task{}

	json.Unmarshal(bodyBytes, &taskToCreate)

	taskToCreate.UserID = userId

	if taskToCreate.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "task name is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskAnswer, err := db.AddTask(ctx, taskToCreate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	jsn, err := json.Marshal(taskAnswer)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, string(jsn))

}

func (db *Db) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "id is must be an integer")
		return
	}

	userId, err := AuthUser(w, r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := db.GetTaskById(ctx, id, userId)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	jsn, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, string(jsn))
}

func (db *Db) GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx, cancer := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancer()

	userId, err := AuthUser(w, r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	resp, err := db.GetAllTasks(ctx, userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	jsn, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, string(jsn))
}

func (db *Db) Delete(w http.ResponseWriter, r *http.Request) {
	userId, err := AuthUser(w, r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "id is must be an integer")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := db.DeleteTaskById(ctx, id, userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	jsn, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, string(jsn))

}

func (db *Db) Update(w http.ResponseWriter, r *http.Request) {
	userId, err := AuthUser(w, r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "id is must be an integer")
		return
	}

	body := r.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "body read error")
		return
	}
	taskToUpdate := TaskUpdate{}
	err = json.Unmarshal(bodyBytes, &taskToUpdate)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.UpdateTaskById(ctx, taskToUpdate, id, userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "success update")
}

func (db *Db) Register(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "body read error")
		return
	}
	user := User{}
	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "body parse error")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.RegisterUser(ctx, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "success register")

}
func (db *Db) Login(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "body read error")
		return
	}
	user := User{}
	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "body parse error")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	token, err := db.LoginUser(ctx, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	jsn, err := json.Marshal(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, string(jsn))
}

func (db *Db) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId, err := AuthUser(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.DeleteUserFromDb(ctx, userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	fmt.Fprint(w, "success delete user")

}

func (db *Db) UpdateUserById(w http.ResponseWriter, r *http.Request) {
	userId, err := AuthUser(w, r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	body := r.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "body read error")
		return
	}
	user := User{}
	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "body parse error")
		return
	}
	user.ID = userId
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.UpdateUser(ctx, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "success update user")
}
