package security

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Service описывает сервис работы с менеджерами.
type Service struct {
	pool *pgxpool.Pool
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
