package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
	"github.com/yngus4862/chat/internal/model"
)

type Store struct {
	db *sql.DB
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// conservative pool defaults (small team)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS chat_rooms (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			room_id INTEGER NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}
	for _, q := range queries {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateRoom(ctx context.Context, name string) (model.Room, error) {
	var r model.Room
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO chat_rooms(name) VALUES($1)
		 RETURNING id, name, created_at`,
		name,
	).Scan(&r.ID, &r.Name, &r.CreatedAt)
	return r, err
}

func (s *Store) ListRooms(ctx context.Context) ([]model.Room, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at FROM chat_rooms ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Room
	for rows.Next() {
		var r model.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) CreateMessage(ctx context.Context, roomID int64, content string) (model.Message, error) {
	if roomID <= 0 {
		return model.Message{}, errors.New("invalid roomID")
	}
	var m model.Message
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO messages(room_id, content) VALUES($1, $2)
		 RETURNING id, room_id, content, created_at`,
		roomID, content,
	).Scan(&m.ID, &m.RoomID, &m.Content, &m.CreatedAt)
	return m, err
}

func (s *Store) ListMessages(ctx context.Context, roomID int64, limit int) ([]model.Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, room_id, content, created_at
		 FROM messages
		 WHERE room_id=$1
		 ORDER BY id DESC
		 LIMIT $2`,
		roomID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}