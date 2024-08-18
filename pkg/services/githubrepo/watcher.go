package githubrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielboakye/github-repo-stats/pkg/db/postgres"
	"github.com/danielboakye/github-repo-stats/pkg/repository"
	"github.com/google/uuid"
)

const (
	githubAPIURL                   = "https://api.github.com"
	githubCommitsMaxRecordsPerPage = 100

	defaultBackoffDuration = 1 * time.Minute

	// ISODateFormat represents ISO 8601 format: YYYY-MM-DDTHH:MM:SSZ
	ISODateFormat = "2006-01-02T15:04:05Z"
)

// GithubRepositoryResponse represents repository http response from github api we are interested in
type GithubRepositoryResponse struct {
	Description      string `json:"description"`
	URL              string `json:"html_url"`
	Language         string `json:"language"`
	ForksCount       int    `json:"forks_count"`
	StarsCount       int    `json:"stargazers_count"`
	OpenIssuesCount  int    `json:"open_issues"`
	SubscribersCount int    `json:"subscribers_count"`
}

// GithubCommitAuthor represents author in GithubCommitDetails
type GithubCommitAuthor struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

// GithubCommitDetails represents commit details in GithubCommitResponse
type GithubCommitDetails struct {
	Message string             `json:"message"`
	Author  GithubCommitAuthor `json:"author"`
}

// GithubCommitResponse represents commit http response
type GithubCommitResponse struct {
	SHA    string              `json:"sha"`
	Commit GithubCommitDetails `json:"commit"`
	URL    string              `json:"html_url"`
}

var (
	// ErrRateLimitReached represents rate limit reached error
	ErrRateLimitReached = fmt.Errorf("rate limit error")
)

// Start fires of listeners for the service
func (s *Service) Start(ctx context.Context) error {
	go s.StartNewReposListener(ctx)
	go s.StartReposWatcher(ctx)

	return nil
}

// StartNewReposListener starts a listens for new repositories and initiates a watch on it
func (s *Service) StartNewReposListener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case repoName := <-s.newRepo:
			go func(repoName string) {
				innerCtx, cancel := context.WithCancel(ctx)
				defer cancel()

				githubRepo, err := s.repo.GetRepositoryByName(innerCtx, repoName)
				if err != nil {
					s.logger.Error("error-getting-repo",
						slog.String("repoName", repoName),
						slog.String("error", err.Error()),
					)
					return
				}

				if err := s.trackRepo(innerCtx, &githubRepo); err != nil {
					s.logger.Error("error-tracking-repo",
						slog.String("repoName", repoName),
						slog.String("error", err.Error()),
					)
					return
				}
			}(repoName)
		}
	}
}

// StartReposWatcher starts the water for pulling commits and repo information
func (s *Service) StartReposWatcher(ctx context.Context) {
	if err := s.trackAllRepos(ctx); err != nil {
		s.logger.Error("error-tracking-repos",
			slog.String("error", err.Error()),
		)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Hour):
			if err := s.trackAllRepos(ctx); err != nil {
				s.logger.Error("error-tracking-repos",
					slog.String("error", err.Error()),
				)
			}
		}
	}
}

func (s *Service) trackRepo(ctx context.Context, repo *repository.GithubRepository) error {
	// update repo meta data
	if err := s.updateRepoInformation(ctx, repo); err != nil {
		return fmt.Errorf("failed tracking repo metadata: %w", err)
	}

	// load commits
	if err := s.trackCommits(ctx, repo); err != nil {
		return fmt.Errorf("failed tracking commits: %w", err)
	}

	return nil
}

func (s *Service) trackAllRepos(ctx context.Context) error {
	s.logger.Info("tracking-all-repos")
	// get names of all repos from db
	repos, err := s.repo.GetRepositories(ctx)
	if errors.Is(err, postgres.ErrRecordNotFound) {
		s.logger.Warn("no-repos-to-watch")
		return nil
	}
	if err != nil {
		s.logger.Error("error-retrieving-repos",
			slog.String("error", err.Error()),
		)
		// send some sort of alert
		return nil
	}

	for _, repo := range repos {
		if err := s.trackRepo(ctx, repo); err != nil {
			s.logger.Error("error-tracking-repo",
				slog.String("repoID", repo.ID),
				slog.String("repoName", repo.RepositoryName),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

func (s *Service) fetchRepoWithRetry(ctx context.Context, repoName string) (GithubRepositoryResponse, error) {
	var (
		ghRepo GithubRepositoryResponse
		err    error
	)
	backoffDuration := defaultBackoffDuration

	for {
		ghRepo, err = s.fetchRepoData(ctx, repoName)
		if errors.Is(err, ErrRateLimitReached) {

			backoffDuration *= 2

			s.logger.Warn("rate-limit-reached:retrying-after-backoff",
				slog.String("repoName", repoName),
				slog.String("backoffDuration", backoffDuration.String()),
			)

			select {
			case <-ctx.Done():
				return ghRepo, ctx.Err()
			case <-time.After(backoffDuration):
				// retry after the backoff period
				continue
			}
		}
		if err != nil {
			return ghRepo, fmt.Errorf("error fetching repo metadata: %w", err)
		}

		break
	}

	return ghRepo, nil
}

func (s *Service) updateRepoInformation(ctx context.Context, gr *repository.GithubRepository) error {
	ghRepo, err := s.fetchRepoWithRetry(ctx, gr.RepositoryName)
	if err != nil {
		return fmt.Errorf("error fetching commits with retry mechanism: %w", err)
	}

	gr.Description = &ghRepo.Description
	gr.URL = &ghRepo.URL
	gr.Language = &ghRepo.Language
	gr.ForksCount = ghRepo.ForksCount
	gr.StarsCount = ghRepo.StarsCount
	gr.OpenIssuesCount = ghRepo.OpenIssuesCount
	gr.WatchersCount = ghRepo.SubscribersCount

	if err := s.repo.UpdateRepository(ctx, gr); err != nil {
		return fmt.Errorf("failed to update repo (%s): %w", gr.ID, err)
	}

	return nil
}

func (s *Service) fetchRepoData(ctx context.Context, repositoryName string) (GithubRepositoryResponse, error) {
	var repo GithubRepositoryResponse
	url := fmt.Sprintf("%s/repos/%s", githubAPIURL, repositoryName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return repo, fmt.Errorf("error fetching from github: %w", err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "go-github-fetcher")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return repo, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return repo, ErrRateLimitReached
	}

	if resp.StatusCode != http.StatusOK {
		return repo, fmt.Errorf("failed to fetch commits: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return repo, err
	}

	return repo, nil
}

func (s *Service) trackCommits(ctx context.Context, repo *repository.GithubRepository) error {
	page := 1
	backoffDuration := defaultBackoffDuration
	var since *time.Time
	// set since to default in flags
	if !s.commitSinceDate.IsZero() {
		since = &s.commitSinceDate
	}
	// if a first set of commits have been pulled. update since to last commit date
	// so previous set of commits are ignored
	if repo.CommitLastPulledTime != nil {
		since = repo.CommitLastPulledTime
		s.logger.Debug("fetching-commits-with-updated-since", slog.String("since", since.Format(ISODateFormat)))
	}

	for {
		numberProcessed, err := s.processUntrackedCommits(ctx, repo.ID, repo.RepositoryName, page, since)
		if errors.Is(err, ErrRateLimitReached) {

			backoffDuration *= 2

			s.logger.Warn("rate-limit-reached:retrying-after-backoff",
				slog.String("repoName", repo.RepositoryName),
				slog.Int("page", page),
				slog.String("since", since.Format(ISODateFormat)),
				slog.String("backoffDuration", backoffDuration.String()),
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoffDuration):
				// retry after the backoff period
				continue
			}
		}
		if err != nil {
			return fmt.Errorf("error fetching commits: %w", err)
		}
		if numberProcessed < githubCommitsMaxRecordsPerPage {
			return nil
		}

		page = page + 1
	}
}

func (s *Service) processUntrackedCommits(ctx context.Context, repoID, repoName string, page int, since *time.Time) (int, error) {
	var numberProcessed int
	url := fmt.Sprintf("%s/repos/%s/commits?page=%d&per_page=%d", githubAPIURL, repoName, page, githubCommitsMaxRecordsPerPage)
	if since != nil {
		url += fmt.Sprintf("&since=%s", since.Format(ISODateFormat))
	}

	s.logger.Info("fetch-commits", slog.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return numberProcessed, fmt.Errorf("error fetching from github: %w", err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "go-github-fetcher")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return numberProcessed, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return numberProcessed, ErrRateLimitReached
	}

	if resp.StatusCode != http.StatusOK {
		return numberProcessed, fmt.Errorf("failed to fetch commits: %s", resp.Status)
	}

	var commits []*GithubCommitResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&commits)
	if err != nil {
		return numberProcessed, err
	}

	for _, commit := range commits {
		if err := s.repo.SaveCommit(ctx, repository.GithubCommit{
			ID:           uuid.New().String(),
			RepositoryID: repoID,
			CommitHash:   commit.SHA,
			Message:      commit.Commit.Message,
			AuthorName:   commit.Commit.Author.Name,
			AuthorEmail:  commit.Commit.Author.Email,
			Date:         commit.Commit.Author.Date,
			URL:          commit.URL,
		}); err != nil {
			return numberProcessed, fmt.Errorf("failed to save new commit: %w", err)
		}
	}

	if len(commits) > 0 {
		lastSyncTime := commits[len(commits)-1].Commit.Author.Date
		if err := s.repo.UpdateCommitLastSyncTime(ctx, repoID, lastSyncTime); err != nil {
			return numberProcessed, fmt.Errorf("failed to update last sync time: %w", err)
		}
	}

	return len(commits), nil
}
