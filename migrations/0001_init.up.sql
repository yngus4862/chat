CREATE TABLE IF NOT EXISTS chat_rooms (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(200) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS messages (
  id BIGSERIAL PRIMARY KEY,
  room_id BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
  content TEXT NOT NULL,
  source VARCHAR(32) NOT NULL DEFAULT 'rest',
  client_msg_id VARCHAR(128) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (room_id, client_msg_id)
);

CREATE INDEX IF NOT EXISTS idx_messages_room_id_id_desc ON messages(room_id, id DESC);