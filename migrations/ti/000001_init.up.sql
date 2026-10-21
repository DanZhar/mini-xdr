CREATE TABLE indicators (
    id          BIGSERIAL PRIMARY KEY,
    type        TEXT NOT NULL CHECK (type IN ('ip', 'domain', 'url', 'md5', 'sha256')),
    value       TEXT NOT NULL,
    threat      TEXT,
    confidence  SMALLINT NOT NULL DEFAULT 50 CHECK (confidence BETWEEN 0 AND 100),
    tags        TEXT[] NOT NULL DEFAULT '{}',
    first_seen  TIMESTAMPTZ NOT NULL,
    last_seen   TIMESTAMPTZ NOT NULL,
    UNIQUE (type, value)
);

CREATE TABLE indicator_sources (
    indicator_id BIGINT NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
    source       TEXT NOT NULL,
    seen_at      TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (indicator_id, source)
);

CREATE TABLE feeds (
    name          TEXT PRIMARY KEY,
    last_fetch_at TIMESTAMPTZ,
    last_status   TEXT,
    last_count    INT
);