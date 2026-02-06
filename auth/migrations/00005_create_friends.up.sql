CREATE TABLE friends (
    ID serial PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(255) NOT NULL DEFAULT 'pending',                 -- Есть три возможных статуса: pending, accepted, rejected
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, friend_id),
    CONSTRAINT no_self_friend CHECK(user_id != friend_id)
);

-- Добавим индексы
CREATE INDEX idx_friends_user_id ON friends(user_id);
CREATE INDEX idx_friends_friend_id on friends(friend_id);
CREATE INDEX idx_friends_status ON friends(status);