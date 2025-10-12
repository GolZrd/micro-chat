CREATE TABLE role_permissions (
    role VARCHAR(64)  NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    PRIMARY KEY (role, endpoint)
);

-- Применим индексы
CREATE INDEX role_permissions_endpoint_idx ON role_permissions (endpoint);