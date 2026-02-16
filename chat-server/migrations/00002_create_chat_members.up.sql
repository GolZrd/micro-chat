CREATE TABLE chat_members (
    ID SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL REFERENCES chats(ID) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    role VARCHAR(255) NOT NULL DEFAULT 'member', -- На будущее (админ, создатель, участник)
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, username)
);

CREATE INDEX idx_chat_members_chat_id ON chat_members(chat_id);
CREATE INDEX idx_chat_members_user_id ON chat_members(user_id);
CREATE INDEX idx_chat_members_username ON chat_members(username);

-- Составной индекс для поиска личных чатов
CREATE INDEX idx_chat_members_chat_user ON chat_members(chat_id, user_id);