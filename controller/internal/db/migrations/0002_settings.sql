-- +migrate Up
CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS enroll_tokens_view AS SELECT 1;
DROP TABLE IF EXISTS enroll_tokens_view;
