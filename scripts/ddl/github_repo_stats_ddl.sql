CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE repository (
    id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    repository_name VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(255),
    language VARCHAR(100),
    forks_count INT DEFAULT 0,
    stars_count INT DEFAULT 0,
    open_issues_count INT DEFAULT 0,
    watchers_count INT DEFAULT 0,
    commit_last_pulled_time TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP,
    UNIQUE (repository_name)
);

CREATE TABLE commits (
    id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    commit_hash VARCHAR(100) NOT NULL,
    repository_id uuid REFERENCES repository(id) ON DELETE CASCADE,
    commit_message TEXT,
    author_name VARCHAR(255),
    author_email VARCHAR(255),
    commit_date TIMESTAMP NOT NULL,
    commit_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    UNIQUE (commit_hash, repository_id)
);

-- Index on repository lookup on name
CREATE INDEX idx_repository_name ON repository(repository_name);

-- Index for loading commits on repository_id FK
CREATE INDEX idx_commits_repository_id ON commits(repository_id);

-- Index for efficiently retrieving the top N commit authors
CREATE INDEX idx_commits_author_name ON commits(author_name);
