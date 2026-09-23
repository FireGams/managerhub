-- +migrate Up
CREATE TABLE users (
    id            UUID PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'viewer',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE enroll_tokens (
    id         UUID PRIMARY KEY,
    token_hash TEXT NOT NULL UNIQUE,
    label      TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE nodes (
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    hostname      TEXT NOT NULL DEFAULT '',
    os            TEXT NOT NULL DEFAULT '',
    arch          TEXT NOT NULL DEFAULT '',
    ip            TEXT NOT NULL DEFAULT '',
    agent_version TEXT NOT NULL DEFAULT '',
    token_hash    TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'offline',
    last_seen_at  TIMESTAMPTZ,
    tags          TEXT[] NOT NULL DEFAULT '{}',
    capabilities  JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX nodes_status_idx ON nodes (status);
CREATE INDEX nodes_name_idx ON nodes (name);

CREATE TABLE node_metrics (
    id         BIGSERIAL PRIMARY KEY,
    node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    ts         TIMESTAMPTZ NOT NULL DEFAULT now(),
    cpu_pct    DOUBLE PRECISION NOT NULL DEFAULT 0,
    cpu_cores  INT NOT NULL DEFAULT 0,
    load1      DOUBLE PRECISION NOT NULL DEFAULT 0,
    ram_used   BIGINT NOT NULL DEFAULT 0,
    ram_total  BIGINT NOT NULL DEFAULT 0,
    disk_used  BIGINT NOT NULL DEFAULT 0,
    disk_total BIGINT NOT NULL DEFAULT 0,
    uptime_sec BIGINT NOT NULL DEFAULT 0,
    gpus       JSONB NOT NULL DEFAULT '[]'
);
CREATE INDEX node_metrics_ts_idx ON node_metrics (ts);
CREATE INDEX node_metrics_node_ts_idx ON node_metrics (node_id, ts DESC);

CREATE TABLE node_events (
    id      BIGSERIAL PRIMARY KEY,
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    ts      TIMESTAMPTZ NOT NULL DEFAULT now(),
    kind    TEXT NOT NULL,
    detail  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX node_events_node_ts_idx ON node_events (node_id, ts DESC);

CREATE TABLE jobs (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL DEFAULT 'shell',
    node_id     UUID REFERENCES nodes(id) ON DELETE SET NULL,
    command     TEXT NOT NULL,
    args        TEXT[] NOT NULL DEFAULT '{}',
    workdir     TEXT NOT NULL DEFAULT '',
    env         JSONB NOT NULL DEFAULT '{}',
    timeout_sec INT NOT NULL DEFAULT 3600,
    status      TEXT NOT NULL DEFAULT 'queued',
    exit_code   INT,
    error       TEXT NOT NULL DEFAULT '',
    stdout      TEXT NOT NULL DEFAULT '',
    stderr      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);
CREATE INDEX jobs_status_idx ON jobs (status);
CREATE INDEX jobs_node_idx ON jobs (node_id);
CREATE INDEX jobs_created_idx ON jobs (created_at DESC);

CREATE TABLE schedules (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    cron_expr   TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    job_name    TEXT NOT NULL,
    job_type    TEXT NOT NULL DEFAULT 'shell',
    command     TEXT NOT NULL,
    args        TEXT[] NOT NULL DEFAULT '{}',
    workdir     TEXT NOT NULL DEFAULT '',
    env         JSONB NOT NULL DEFAULT '{}',
    timeout_sec INT NOT NULL DEFAULT 3600,
    node_id     UUID REFERENCES nodes(id) ON DELETE SET NULL,
    selector    JSONB NOT NULL DEFAULT '{}',
    last_run_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE schedule_runs (
    id          BIGSERIAL PRIMARY KEY,
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    job_id      UUID REFERENCES jobs(id) ON DELETE SET NULL,
    ts          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX schedule_runs_sched_idx ON schedule_runs (schedule_id, ts DESC);

CREATE TABLE audit_log (
    id      BIGSERIAL PRIMARY KEY,
    ts      TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor   TEXT NOT NULL,
    action  TEXT NOT NULL,
    target  TEXT NOT NULL DEFAULT '',
    detail  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX audit_ts_idx ON audit_log (ts DESC);

CREATE TABLE terminal_sessions (
    id         UUID PRIMARY KEY,
    node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    user_name  TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at   TIMESTAMPTZ,
    reason     TEXT NOT NULL DEFAULT ''
);
