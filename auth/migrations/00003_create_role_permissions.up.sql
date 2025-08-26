CREATE TABLE role_permissions (
    role VARCHAR(64)  NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    PRIMARY KEY (role, endpoint)
);