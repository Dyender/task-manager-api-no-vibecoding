package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	cr "golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (db *Db) RegisterUserDB(ctx context.Context, user *User) error {
	hash, err := cr.GenerateFromPassword([]byte(user.Password), cr.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.dbConnect.Query(
		ctx,
		`insert into users (login, password, role) values ($1, $2, $3)`,
		user.Login,
		hash,
		user.Role,
	)
	return err
}

func (db *Db) LoginUserDB(ctx context.Context, user *User) (string, error) {
	userID := 0
	role := ""
	hash := ""

	err := db.dbConnect.QueryRow(
		ctx,
		`select id, password, role from users where login = $1`,
		user.Login,
	).Scan(&userID, &hash, &role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errors.New("invalid login or password")
	}
	if err != nil {
		return "", err
	}

	err = cr.CompareHashAndPassword([]byte(hash), []byte(user.Password))
	if err != nil {
		return "", errors.New("invalid login or password")
	}

	jwtClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24 * 15).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	mySecretToken := os.Getenv("MY_SECRET_TOKEN")
	if mySecretToken == "" {
		mySecretToken = "secret"
	}
	tokenString, err := token.SignedString([]byte(mySecretToken))
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func (db *Db) UpdateUserDB(ctx context.Context, user *User) error {
	hash, err := cr.GenerateFromPassword([]byte(user.Password), cr.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	_, err = db.dbConnect.Query(
		ctx,
		`update users set login = $1, password = $2 where id = $3`,
		user.Login,
		user.Password,
		user.ID,
	)
	return err
}

func (db *Db) DeleteUserDB(ctx context.Context, id int) error {
	//добавить удаление всех тасков юзера по его айди
	_, err := db.dbConnect.Query(
		ctx,
		`delete from tasks where userID = $1`,
		id,
	)
	if err != nil {
		return err
	}
	_, err = db.dbConnect.Query(
		ctx,
		`delete from users where id = $1`,
		id,
	)
	return err
}

func (db *Db) GetAllUsersDB(ctx context.Context) ([]*User, error) {
	users := []*User{}

	rows, err := db.dbConnect.Query(
		ctx,
		`select * from users`,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		user := &User{}
		err = rows.Scan(&user.ID, &user.Login, &user.Password, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)

	}
	return users, nil
}

func (db *Db) GetUserByIdDB(ctx context.Context, id int) (*User, error) {
	user := &User{}
	err := db.dbConnect.QueryRow(
		ctx,
		`select * from users where id = $1`,
		id,
	).Scan(&user.ID, &user.Login, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return user, nil
}
