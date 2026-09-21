-- +goose Up
CREATE TABLE `groups` (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    billing_multiplier DOUBLE NOT NULL DEFAULT 1 CHECK (billing_multiplier >= 0),
    visible_other_group BOOLEAN NOT NULL DEFAULT FALSE,
    description VARCHAR(1000) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- +goose Down
DROP TABLE `groups`;
