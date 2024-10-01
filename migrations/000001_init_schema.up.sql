CREATE TYPE provider_type AS ENUM ('email', 'google', 'apple', 'firebase', 'github', 'twitter', 'linkedin', 'microsoft', 'gitlab', 'bitbucket', 'keycloak', 'oidc', 'oauth', 'custom');

CREATE TYPE action_type AS ENUM ('CREATE', 'READ', 'UPDATE', 'DELETE', 'LIST');

CREATE TYPE resource_type AS ENUM('USER', 'ROLE', 'PERMISSION', 'PRODUCT', 'ORDER');


CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(100) NOT NULL UNIQUE CHECK (length(username) >= 2),
                       password_hash VARCHAR(255) NOT NULL,
                       email VARCHAR(100) NOT NULL UNIQUE CHECK (email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+\.[A-Za-z]+$'),
                       phone VARCHAR(20) DEFAULT '' NOT NULL,
                       firebase_uid VARCHAR(255) UNIQUE,
                       provider provider_type NOT NULL DEFAULT 'email',
                       display_name VARCHAR(255),
                       photo_url VARCHAR(512),
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);


CREATE TABLE roles (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255) NOT NULL UNIQUE,
                       description VARCHAR(255),
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
                             id SERIAL PRIMARY KEY,
                             name VARCHAR(255) NOT NULL UNIQUE,
                             description VARCHAR(255),
                             resource resource_type NOT NULL,
                             action action_type NOT NULL,
                             created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                             updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
                            user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
                            role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
                            PRIMARY KEY (user_id, role_id)
);

CREATE TABLE role_permissions (
                                  role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
                                  permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
                                  PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_roles_name ON roles (name);
CREATE INDEX idx_permissions_resource ON permissions (resource);
CREATE INDEX idx_permissions_action ON permissions (action);
CREATE INDEX idx_user_roles_user_id ON user_roles (user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions (role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);


-- 插入用戶
INSERT INTO users (username, password_hash, email) VALUES
                                                       ('admin', '$2a$10$Gqpa6ukRxP.XxwCLjNRrweuWPgrw0XIIo5xi.a8XUpVcndvIgWPlW', 'admin@example.com'),
                                                       ('user1', '$2a$10$Gqpa6ukRxP.XxwCLjNRrweuWPgrw0XIIo5xi.a8XUpVcndvIgWPlW', 'user1@example.com'),
                                                       ('user2', '$2a$10$Gqpa6ukRxP.XxwCLjNRrweuWPgrw0XIIo5xi.a8XUpVcndvIgWPlW', 'user2@example.com');

-- 插入角色
INSERT INTO roles (name, description) VALUES
                                          ('admin', 'Administrator with full access'),
                                          ('manager', 'Manager with elevated privileges'),
                                          ('user', 'Regular user with basic access');

-- 插入權限
INSERT INTO permissions (name, description, resource, action) VALUES
                                                                  ('create_user', 'Create a new user', 'USER', 'CREATE'),
                                                                  ('read_user', 'Read user information', 'USER', 'READ'),
                                                                  ('update_user', 'Update user information', 'USER', 'UPDATE'),
                                                                  ('delete_user', 'Delete a user', 'USER', 'DELETE'),
                                                                  ('list_users', 'List all users', 'USER', 'LIST'),
                                                                  ('manage_roles', 'Manage roles', 'ROLE', 'UPDATE'),
                                                                  ('manage_permissions', 'Manage permissions', 'PERMISSION', 'UPDATE'),
                                                                  ('create_product', 'Create a new product', 'PRODUCT', 'CREATE'),
                                                                  ('read_product', 'Read product information', 'PRODUCT', 'READ'),
                                                                  ('update_product', 'Update product information', 'PRODUCT', 'UPDATE'),
                                                                  ('delete_product', 'Delete a product', 'PRODUCT', 'DELETE'),
                                                                  ('list_products', 'List all products', 'PRODUCT', 'LIST');

-- 分配角色給用戶
INSERT INTO user_roles (user_id, role_id) VALUES
                                              ((SELECT id FROM users WHERE username = 'admin'), (SELECT id FROM roles WHERE name = 'admin')),
                                              ((SELECT id FROM users WHERE username = 'user1'), (SELECT id FROM roles WHERE name = 'manager')),
                                              ((SELECT id FROM users WHERE username = 'user2'), (SELECT id FROM roles WHERE name = 'user'));

-- 分配權限給角色
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'manager' AND p.name NOT IN ('delete_user', 'manage_roles', 'manage_permissions');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN ('read_user', 'read_product', 'list_products');
