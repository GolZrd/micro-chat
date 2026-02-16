CREATE TABLE chats (
    ID BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_direct BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chats_is_direct ON chats(is_direct);
CREATE INDEX idx_chats_created_at ON chats(created_at DESC);