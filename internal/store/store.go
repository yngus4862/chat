package store

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.pool == nil {
		return errors.New("store not initialized")
	}
	return s.pool.Ping(ctx)
}

func (s *Store) CreateRoom(ctx context.Context, name string) (Room, error) {
	var r Room
	err := s.pool.QueryRow(ctx,
		`INSERT INTO chat_rooms(name) VALUES($1)
			 RETURNING id, name, created_at`,
		name,
	).Scan(&r.ID, &r.Name, &r.CreatedAt)
	return r, err
}

func (s *Store) ListRooms(ctx context.Context, limit int) ([]Room, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `SELECT id, name, created_at FROM chat_rooms ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Room, 0, limit)
	for rows.Next() {
		var r Room
		if err := rows.Scan(&r.ID, &r.Name, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) ListMessages(ctx context.Context, roomID int64, cursor int64, limit int) ([]Message, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var rows pgx.Rows
	var err error
	if cursor > 0 {
		rows, err = s.pool.Query(ctx,
			`SELECT id, room_id, content, source, client_msg_id, created_at
				 FROM messages
				 WHERE room_id=$1 AND id < $2
				 ORDER BY id DESC
				 LIMIT $3`,
			roomID, cursor, limit,
		)
	} else {
		rows, err = s.pool.Query(ctx,
			`SELECT id, room_id, content, source, client_msg_id, created_at
				 FROM messages
				 WHERE room_id=$1
				 ORDER BY id DESC
				 LIMIT $2`,
			roomID, limit,
		)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]Message, 0, limit)
	var nextCursor int64 = 0
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.Content, &m.Source, &m.ClientMsgID, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
		nextCursor = m.ID
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, nextCursor, nil
}

func (s *Store) CreateMessage(ctx context.Context, roomID int64, content, source, clientMsgID string) (Message, error) {
	if clientMsgID == "" {
		clientMsgID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	var m Message
	err := s.pool.QueryRow(ctx,
		`INSERT INTO messages(room_id, content, source, client_msg_id)
			 VALUES($1,$2,$3,$4)
			 ON CONFLICT (room_id, client_msg_id)
			 DO UPDATE SET content = messages.content
			 RETURNING id, room_id, content, source, client_msg_id, created_at`,
		roomID, content, source, clientMsgID,
	).Scan(&m.ID, &m.RoomID, &m.Content, &m.Source, &m.ClientMsgID, &m.CreatedAt)
	return m, err
}
