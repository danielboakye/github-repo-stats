package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/danielboakye/github-repo-stats/pkg/repository"
	"github.com/google/uuid"
)

// Repository represents the postgres implementation of repository.Repository
type Repository struct {
	db *sql.DB
}

// NewRepository initiates a new postgres repository
func NewRepository(db *sql.DB) repository.Repository {
	return &Repository{db: db}
}

// ErrRecordNotFound represents no record found in postgres datastore
var ErrRecordNotFound = sql.ErrNoRows

// GetRepositories implements repository.Repository
func (p *Repository) GetRepositories(ctx context.Context) ([]*repository.GithubRepository, error) {
	var repositories []*repository.GithubRepository
	query := `
        SELECT id, repository_name, commit_last_pulled_time 
        FROM repository
    `
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		repo := &repository.GithubRepository{}
		err := rows.Scan(
			&repo.ID,
			&repo.RepositoryName,
			&repo.CommitLastPulledTime,
		)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, repo)
	}

	err = rows.Err()
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}

	return repositories, nil
}

// GetRepositoryByName implements repository.Repository
func (p *Repository) GetRepositoryByName(ctx context.Context, name string) (repository.GithubRepository, error) {
	var repo repository.GithubRepository
	query := `
        SELECT id, repository_name, commit_last_pulled_time, description, url, language, forks_count, stars_count, open_issues_count, watchers_count, created_at, updated_at
        FROM repository
        WHERE repository_name = $1
    `
	err := p.db.QueryRowContext(ctx, query, name).Scan(
		&repo.ID,
		&repo.RepositoryName,
		&repo.CommitLastPulledTime,
		&repo.Description,
		&repo.URL,
		&repo.Language,
		&repo.ForksCount,
		&repo.StarsCount,
		&repo.OpenIssuesCount,
		&repo.WatchersCount,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return repo, ErrRecordNotFound
	}
	if err != nil {
		return repo, err
	}

	return repo, nil
}

// CreateRepository implements repository.Repository
func (p *Repository) CreateRepository(ctx context.Context, repoName string) (string, error) {
	repoID := uuid.New().String()
	query := `
		INSERT INTO repository (id, repository_name)
		VALUES ($1,$2)
		`
	_, err := p.db.ExecContext(ctx, query, repoID, repoName)
	if err != nil {
		return repoID, fmt.Errorf("could not insert repository: %w", err)
	}

	return repoID, nil
}

// UpdateRepository implements repository.Repository
func (p *Repository) UpdateRepository(ctx context.Context, repo *repository.GithubRepository) error {
	query := `
        UPDATE repository
        SET 
            description = $1,
            url = $2,
            language = $3,
            forks_count = $4,
            stars_count = $5,
            open_issues_count = $6,
            watchers_count = $7,
            updated_at = $8
        WHERE id = $9
    `
	_, err := p.db.ExecContext(ctx, query,
		repo.Description,
		repo.URL,
		repo.Language,
		repo.ForksCount,
		repo.StarsCount,
		repo.OpenIssuesCount,
		repo.WatchersCount,
		time.Now(),
		repo.ID,
	)
	if err != nil {
		return fmt.Errorf("could not update repository: %w", err)
	}

	return nil
}

// UpdateCommitLastSyncTime implements repository.Repository
func (p *Repository) UpdateCommitLastSyncTime(ctx context.Context, repoID string, syncTime time.Time) error {
	query := `
	UPDATE repository
	SET 
		commit_last_pulled_time = $1
	WHERE id = $2
	`
	_, err := p.db.ExecContext(ctx, query,
		syncTime,
		repoID,
	)
	if err != nil {
		return fmt.Errorf("could not update repository: %w", err)
	}

	return nil
}

// SaveCommit implements repository.Repository
func (p *Repository) SaveCommit(ctx context.Context, commit repository.GithubCommit) error {
	query := `
		INSERT INTO commits (commit_hash, repository_id, commit_message, author_name, author_email, commit_date, commit_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
    	ON CONFLICT (commit_hash, repository_id) DO NOTHING`
	_, err := p.db.ExecContext(ctx, query,
		commit.CommitHash,
		commit.RepositoryID,
		commit.Message,
		commit.AuthorName,
		commit.AuthorEmail,
		commit.Date,
		commit.URL,
	)
	if err != nil {
		return fmt.Errorf("could not insert repository: %w", err)
	}

	return nil
}

// GetCommitsByRepository implements repository.Repository
func (p *Repository) GetCommitsByRepository(ctx context.Context, repoID string, limit, offset int) ([]*repository.GithubCommit, error) {
	var commits []*repository.GithubCommit
	query := `
        SELECT commit_hash, commit_message, author_name, author_email, commit_date, commit_url 
        FROM commits
		WHERE repository_id=$1
		ORDER BY commit_date DESC
		LIMIT $2 OFFSET $3
    `
	rows, err := p.db.QueryContext(ctx, query, repoID, limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		commit := &repository.GithubCommit{}
		err := rows.Scan(
			&commit.CommitHash,
			&commit.Message,
			&commit.AuthorName,
			&commit.AuthorEmail,
			&commit.Date,
			&commit.URL,
		)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	err = rows.Err()
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}

	return commits, err
}

// GetLeaderBoard implements repository.Repository
func (p *Repository) GetLeaderBoard(ctx context.Context, limit int) ([]repository.CommitStats, error) {
	var leaderboard []repository.CommitStats
	query := `
	SELECT author_name, COUNT(id) as commit_count
	FROM commits
	GROUP BY author_name
	ORDER BY commit_count DESC
	LIMIT $1
    `
	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		stat := repository.CommitStats{}
		err := rows.Scan(
			&stat.AuthorName,
			&stat.CommitCount,
		)
		if err != nil {
			return nil, err
		}

		leaderboard = append(leaderboard, stat)
	}

	err = rows.Err()
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}

	return leaderboard, err
}
