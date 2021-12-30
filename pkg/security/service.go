package security

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// ErrNoSuchUser если пользователь не найден
var ErrNoSuchUser = errors.New("no such user")

// ErrInvalidPassword если пароль не верный
var ErrInvalidPassword = errors.New("invalid password")

// ErrInternal возвращается когда произошла внутренная ошибка.
var ErrInternal = errors.New("internal error")

// ErrExpiredToken возвращается когда чувачок исчерпал свой токен
var ErrExpiredToken = errors.New("Token is expired")

// Service описывает сервис работы с менеджерами.
type Service struct {
	pool *pgxpool.Pool
}

type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// NewService создаёт сервис
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Managers представляет информацию о менеджере.
type Managers struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Login      string    `json:"login"`
	Password   string    `json:"password"`
	Salary     int       `json:"salary"`
	Plan       int       `json:"plan"`
	BossID     int64     `json:"boss_id"`
	Department string    `json:"department"`
	Active     bool      `json:"active"`
	Created    time.Time `json:"created"`
}

// AuthentificateCustomer проводит процедуру аутентификации покупателя,
// возвращая в случае успеха его id.
//	Если пользователь не найден, возвращается ErrNoSuchUser.
//	Если пароль не верен, возвращается ErrInvalidPassword.
//	Если происходит другая ошибка, возвращается ErrInternal.

func (s *Service) AuthentificateCustomer(
	ctx context.Context,
	token string,
) (id int64, err error) {
	err = s.pool.QueryRow(ctx, `SELECT customer_id FROM customers_tokens WHERE token = $1`, token).Scan(&id)

	if err == pgx.ErrNoRows {
		return 0, ErrNoSuchUser
	}
	if err != nil {
		return 0, ErrInternal
	}

	return id, nil
}

func (s Service) AuthForCustomer(
	ctx context.Context,
	token string,
) (id int64, err error) {
	// Время когда исчерпается авторизация
	expiredTime := time.Now()
	nowTimeInSec := expiredTime.UnixNano()
	err = s.pool.QueryRow(ctx, `SELECT customer_id, expire FROM customers_tokens WHERE token = $1`, token).Scan(&id, &expiredTime)
	if err != nil {
		log.Print(err)
		return 0, ErrNoSuchUser
	}

	if nowTimeInSec > expiredTime.UnixNano() {
		return -1, ErrExpiredToken
	}
	return id, nil
}

// Auth - метод авторизации.
func (s *Service) Auth(login, password string) bool {

	log.Print("Go to func Auth")
	log.Print(login, password)

	pass := ""
	ctx := context.Background()
	err := s.pool.QueryRow(ctx, `
		SELECT password FROM managers WHERE login = $1
	`, login).Scan(&pass)

	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}

	if err != nil {
		log.Print(err)
		return false
	}

	log.Print(pass)
	if pass == password {
		return true
	}
	return false
}
