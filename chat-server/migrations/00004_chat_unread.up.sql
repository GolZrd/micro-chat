-- Таблица непрочитанных сообщений
CREATE TABLE chat_unread (
    chat_id      BIGINT NOT NULL REFERENCES chats(ID) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL,
    count        INT NOT NULL DEFAULT 0,
    last_read_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX idx_chat_unread_user ON chat_unread(user_id);
CREATE INDEX idx_chat_unread_chat ON chat_unread(chat_id);
CREATE INDEX idx_chat_unread_user_nonzero ON chat_unread(user_id) WHERE count > 0;

-- Индекс для быстрого получения последних сообщений
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);