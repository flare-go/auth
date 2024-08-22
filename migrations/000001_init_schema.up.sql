-- 創建自定義類型

CREATE TYPE action_type AS ENUM ('CREATE', 'READ', 'UPDATE', 'DELETE', 'LIST');

CREATE TYPE resource_type AS ENUM('USER', 'ROLE', 'PERMISSION', 'PRODUCT', 'ORDER');


CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(100) NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       email VARCHAR(100) NOT NULL UNIQUE CHECK (position('@' in email) > 0),
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE roles (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255) NOT NULL UNIQUE,
                       description TEXT,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
                             id SERIAL PRIMARY KEY,
                             name VARCHAR(255) NOT NULL UNIQUE,
                             description TEXT,
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
