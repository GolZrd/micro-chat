INSERT INTO role_permissions (role, endpoint) VALUES
    -- Chat service - USER
    ('user', '/chat.ChatV1/Create'),
    ('user', '/chat.ChatV1/MyChats'),
    ('user', '/chat.ChatV1/SendMessage'),
    ('user', '/chat.ChatV1/ConnectChat'),
    -- Chat service - ADMIN
    ('admin', '/chat.ChatV1/Create'),
    ('admin', '/chat.ChatV1/MyChats'),
    ('admin', '/chat.ChatV1/SendMessage'),
    ('admin', '/chat.ChatV1/ConnectChat'),
    -- Auth service
    ('user', '/auth.AuthV1/GetUserInfo'),
    ('admin', '/auth.AuthV1/GetUserInfo')
    -- ADMIN
ON CONFLICT (role, endpoint) DO NOTHING;

-- Все остальные методы у нас по умолчанию доступны всем