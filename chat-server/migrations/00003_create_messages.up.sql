CREATE TABLE messages (
    ID BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL REFERENCES chats(ID) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(ID) ON DELETE SET NULL,
    from_username VARCHAR(255) NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text', -- На будущее
    text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_user_id ON messages(user_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);