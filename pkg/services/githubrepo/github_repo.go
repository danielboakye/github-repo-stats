package githubrepo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielboakye/github-repo-stats/pkg/db/postgres"
	"github.com/danielboakye/github-repo-stats/pkg/repository"
)

// Service represents github repos service
type Service struct {
	repo            repository.Repository
	logger          *slog.Logger
	httpClient      *http.Client
	newRepo         chan string
	commitSinceDate time.Time
}

// NewService initiates a new github service manager
func NewService(repo repository.Repository, logger *slog.Logger, commitSinceDate time.Time) *Service {
	return &Service{
		repo:            repo,
		logger:          logger,
		httpClient:      &http.Client{},
		newRepo:         make(chan string, 100),
		commitSinceDate: commitSinceDate,
	}
}

// getRepositoryID returns github repository id from the datastore
// if github repo does not exist in the datastore, it creates it and returns its id
// it add repo to listener queue to trigger watch (pulling of commits and metadata) on the repo
func (s *Service) getRepositoryID(ctx context.Context, repoName string) (string, error) {
	var githubRepoID string
	githubRepo, err := s.repo.GetRepositoryByName(ctx, repoName)
	if errors.Is(err, postgres.ErrRecordNotFound) {
		returnID, rpErr := s.repo.CreateRepository(ctx, repoName)
		if rpErr != nil {
			s.logger.Error("error-creating-repo",
				slog.String("repoName", repoName),
				slog.String("error", err.Error()),
			)
			return githubRepoID, fmt.Errorf("error creating repository: %w", err)
		}

		// send message to channel to trigger loading
		s.newRepo <- repoName

		githubRepoID = returnID
		return githubRepoID, nil
	}
	if err != nil {
		return githubRepoID, fmt.Errorf("failed to get repository (%s) with error: %w", repoName, err)
	}

	return githubRepo.ID, nil
}

// GetCommits loads paginated commits for a github repo
func (s *Service) GetCommits(ctx context.Context, repoName string, limit, offset int) ([]*repository.GithubCommit, error) {
	githubRepoID, err := s.getRepositoryID(ctx, repoName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving repository id: %w", err)
	}
	commits, err := s.repo.GetCommitsByRepository(ctx, githubRepoID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits for repository (%v) with error: %w", githubRepoID, err)
	}

	return commits, nil
}

// GetLeaderBoard loads leaderboard stats
func (s *Service) GetLeaderBoard(ctx context.Context, count int) ([]repository.CommitStats, error) {
	stats, err := s.repo.GetLeaderBoard(ctx, count)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits stats: %w", err)
	}

	return stats, nil
}
