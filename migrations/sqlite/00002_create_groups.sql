-- +goose Up
CREATE TABLE `groups` (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    billing_multiplier REAL NOT NULL DEFAULT 1 CHECK (billing_multiplier >= 0),
    visible_other_group BOOLEAN NOT NULL DEFAULT FALSE,
    description VARCHAR(1000) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE `groups`;
