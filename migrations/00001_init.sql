-- +goose Up

CREATE TABLE roles (
    id   INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

INSERT INTO roles (slug, name) VALUES
    ('frontend',        'Frontend-разработчик'),
    ('backend',         'Backend-разработчик'),
    ('fullstack',       'Fullstack-разработчик'),
    ('project-manager', 'Project Manager'),
    ('product-manager', 'Product Manager'),
    ('ux-ui-designer',  'UX/UI Дизайнер'),
    ('analyst',         'Аналитик'),
    ('logo-designer',   'Лого-дизайнер'),
    ('qa',              'QA'),
    ('devops',          'DevOps');

CREATE TABLE users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tg_id       INTEGER NOT NULL UNIQUE,
    tg_username TEXT    NOT NULL DEFAULT '',
    name        TEXT    NOT NULL DEFAULT '',
    bio         TEXT    NOT NULL DEFAULT '',
    experience  TEXT    NOT NULL DEFAULT '',
    skills      TEXT    NOT NULL DEFAULT '[]',
    onboarded   INTEGER NOT NULL DEFAULT 0,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_roles (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE projects (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    author_id   INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT     NOT NULL,
    description TEXT     NOT NULL DEFAULT '',
    stack       TEXT     NOT NULL DEFAULT '[]',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE project_roles (
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    role_id    INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, role_id)
);

CREATE TABLE responses (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     TEXT    NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, user_id)
);

CREATE TABLE notifications (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT    NOT NULL,
    payload    TEXT    NOT NULL DEFAULT '{}',
    read       INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    token      TEXT     PRIMARY KEY,
    user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

CREATE INDEX idx_users_tg_id ON users(tg_id);
CREATE INDEX idx_projects_author ON projects(author_id);
CREATE INDEX idx_projects_created ON projects(created_at DESC);
CREATE INDEX idx_responses_project ON responses(project_id);
CREATE INDEX idx_responses_user ON responses(user_id);
CREATE INDEX idx_notifications_user ON notifications(user_id, read);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- +goose Down
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS responses;
DROP TABLE IF EXISTS project_roles;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
