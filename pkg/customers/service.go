package customers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ErrNotFound возвращается, когда покупатель не найден.
var ErrNotFound = errors.New("item not found")

// ErrInternal возвращается, когда произошла внутренняя ошибка.
var ErrInternal = errors.New("internal error")

// ErrNoSuchUser если пользователь не найден
var ErrNoSuchUser = errors.New("No such user")

// ErrInvalidPassword если пароль не верный
var ErrInvalidPassword = errors.New("Invalid password")

// ErrExpiredToken возвращается когда чувак исчерпал свой токен
var ErrExpiredToken = errors.New("Token is expired")

// Service описывает сервис работы с покупателями.
type Service struct {
	pool  *pgxpool.Pool
	mu    sync.RWMutex
	items []*Customer
}

// NewService создаёт сервис.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Customer представляет информацию о покупателе.
type Customer struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Phone    string    `json:"phone"`
	Password string    `json:"password"`
	Active   bool      `json:"active"`
	Created  time.Time `json:"created"`
}

//	TokenForCustomer генерирует токен для пользователя.
//	Если пользователь не найден, возвращается ErrNoSuchUser.
//	Если пароль не верен, возвращается ErrInvalidPassword.
//	Если происходит другая ошибка, возвращается ErrInternal.
func (s *Service) TokenForCustomer(
	ctx context.Context,
	phone string,
	password string,
) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `SELECT id, password FROM customers WHERE phone = $1`, phone).Scan(&id, &hash)

	if err == pgx.ErrNoRows {
		return "", ErrInvalidPassword
	}
	if err != nil {
		return "", ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO customers_tokens(token, customer_id) VALUES($1, $2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}

// SaveCustomer сохраняет покупателя с паролем в файле JSON
func (s *Service) SaveCustomer(ctx context.Context, item *Customer) (*Customer, error) {

	if item.ID == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}
		log.Print(hex.EncodeToString(hash))

		res := &Customer{}
		err = s.pool.QueryRow(ctx, `
			INSERT INTO customers(name, phone, password) VALUES ($1, $2, $3)
			RETURNING id, name, phone, password, active, created
			`, item.Name, item.Phone, hash).Scan(&res.ID, &res.Name, &res.Phone, &res.Password, &res.Active, &res.Created)

		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}
		return res, nil
	}
	return nil, ErrInternal
}

// ByID возвращает покупателя по идентификатору.
func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item := &Customer{}

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE id = $1 
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

// All возвращает все данные покупателя.
func (s *Service) All(ctx context.Context) ([]*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// создаём слайс для хранения результатов
	items := make([]*Customer, 0)

	// делаем сам запрос
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, phone, active, created FROM customers
	`)

	// проверяем на ошибки
	if err != nil {
		log.Print(err)
		return nil, err
	}
	// rows нужно закрывать
	defer rows.Close()

	// rows.Next() возвращает true до тех пор, пока дальше есть строки
	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	// в конце нужно проверять общую ошибку
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

// AllActive возвращает все данные активных покупателей.
func (s *Service) AllActive(ctx context.Context) ([]*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// создаём слайс для хранения результатов
	items := make([]*Customer, 0)
	// делаем сам запрос
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE active
	`)
	// проверяем на ошибки
	if err != nil {
		log.Print(err)
		return nil, ErrNotFound
	}
	// rows нужно закрывать
	defer rows.Close()

	// rows.Next() возвращает true до тех пор, пока дальше есть строки
	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	// в конце нужно проверять общую ошибку
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

// Save - создаёт/обновляет покупателя.
func (s *Service) Save(ctx context.Context, item *Customer) (*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := &Customer{}
	if item.ID == 0 {
		err := s.pool.QueryRow(ctx, `
			INSERT INTO customers(name, phone) VALUES ($1, $2) ON CONFLICT (phone) DO UPDATE SET name = excluded.name, active = excluded.active, created = excluded.created
			RETURNING id, name, phone, active, created
		`, item.Name, item.Phone).Scan(&res.ID, &res.Name, &res.Phone, &res.Active, &res.Created)

		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}

		return res, nil
	}

	_, err := s.ByID(ctx, item.ID)
	if err != nil {
		log.Print(err)
		return nil, ErrNotFound
	}
	err = s.pool.QueryRow(ctx, `
		UPDATE customers SET name = $1, phone = $2, active = $3, created = $4 where id = $5 RETURNING id, name, phone, active, created
	`, item.Name, item.Phone, item.Active, item.Created, item.ID).Scan(&res.ID, &res.Name, &res.Phone, &res.Active, &res.Created)

	if err != nil {
		log.Print(err)
		return nil, err
	}

	return res, nil
}

// RemoveByID удаляет пользователя по идентификатору.
func (s *Service) RemoveByID(ctx context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, err := s.ByID(ctx, id)
	if err != nil {
		log.Print(err)
		return ErrNotFound
	}
	_, err = s.pool.Exec(ctx, `
		DELETE FROM customers WHERE id = $1;
	`, id)

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

// BlockByID выставляет статус active в false
func (s *Service) BlockByID(ctx context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, err := s.ByID(ctx, id)
	if err != nil {
		log.Print(err)
		return ErrNotFound
	}

	_, err = s.pool.Exec(ctx, `
	 	UPDATE customers SET active = false WHERE id = $1
	`, id)

	if err != nil {
		log.Print(err)
		return ErrInternal
	}

	return nil
}

// UnBlockByID выставляет статус active в true
func (s *Service) UnBlockByID(ctx context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, err := s.ByID(ctx, id)
	if err != nil {
		log.Print(err)
		return ErrNotFound
	}

	_, err = s.pool.Exec(ctx, `
	 	UPDATE customers SET active = true WHERE id = $1
	`, id)

	if err != nil {
		log.Print(err)
		return ErrInternal
	}

	return nil
}
