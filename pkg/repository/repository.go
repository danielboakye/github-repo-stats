package repository

import (
	"context"
	"time"
)

// Repository represents the interface for all database operations
type Repository interface {
	GetRepositories(ctx context.Context) ([]*GithubRepository, error)
	GetRepositoryByName(ctx context.Context, name string) (GithubRepository, error)
	CreateRepository(ctx context.Context, repoName string) (string, error)
	UpdateRepository(ctx context.Context, repo *GithubRepository) error
	UpdateCommitLastSyncTime(ctx context.Context, repoID string, syncTime time.Time) error
	SaveCommit(ctx context.Context, commit GithubCommit) error
	GetCommitsByRepository(ctx context.Context, repoID string, limit, offset int) ([]*GithubCommit, error)
	GetLeaderBoard(ctx context.Context, limit int) ([]CommitStats, error)
}
