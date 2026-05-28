package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/auth"
	"life-sim/backend/config"
	"life-sim/backend/model"
)

var ErrUserNotFound = errors.New("user not found")

func (s *Store) GetUserByUsername(username string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, display_name, created_at, updated_at FROM users WHERE username = ?`,
		username,
	)
	return scanUser(row)
}

func (s *Store) GetUserByID(id string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, display_name, created_at, updated_at FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var created, updated string
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.CreatedAt = parseDBTime(created)
	u.UpdatedAt = parseDBTime(updated)
	return &u, nil
}

func (s *Store) CreateUser(username, password, displayName string) (*model.User, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	u := &model.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: hash,
		DisplayName:  displayName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if u.DisplayName == "" {
		u.DisplayName = username
	}
	_, err = s.db.Exec(
		`INSERT INTO users (id, username, password_hash, display_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, u.DisplayName, u.CreatedAt, u.UpdatedAt,
	)
	return u, err
}

func (s *Store) UpdateUserPassword(userID, newPassword string) error {
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	now := time.Now()
	res, err := s.db.Exec(
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		hash, now, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *Store) BootstrapUsers(users []config.BootstrapUser) (int, error) {
	created := 0
	for _, bu := range users {
		if bu.Username == "" || bu.Password == "" {
			continue
		}
		_, err := s.GetUserByUsername(bu.Username)
		if err == nil {
			continue
		}
		if !errors.Is(err, ErrUserNotFound) {
			return created, err
		}
		if _, err := s.CreateUser(bu.Username, bu.Password, bu.Username); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}
