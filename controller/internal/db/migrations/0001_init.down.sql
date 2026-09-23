-- +migrate Down
DROP TABLE IF EXISTS terminal_sessions;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS schedule_runs;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS node_events;
DROP TABLE IF EXISTS node_metrics;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS enroll_tokens;
DROP TABLE IF EXISTS users;
