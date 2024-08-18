package repository

import "time"

// GithubRepository represents github repository
type GithubRepository struct {
	ID                   string     `json:"id"`
	RepositoryName       string     `json:"repository_name"` // repository_name is of format {owner}/{repo}
	Description          *string    `json:"description"`
	URL                  *string    `json:"url"`
	Language             *string    `json:"language"`
	ForksCount           int        `json:"forks_count"`
	StarsCount           int        `json:"stars_count"`
	OpenIssuesCount      int        `json:"open_issues_count"`
	WatchersCount        int        `json:"watchers_count"`
	CommitLastPulledTime *time.Time `json:"commit_last_pulled_time"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at"`
}

// GithubCommit represents git commit
type GithubCommit struct {
	ID           string    `json:"id,omitempty"`
	RepositoryID string    `json:"repository_id,omitempty"`
	CommitHash   string    `json:"commit_hash"`
	Message      string    `json:"message"`
	AuthorName   string    `json:"author_name"`
	AuthorEmail  string    `json:"author_email"`
	Date         time.Time `json:"date"`
	URL          string    `json:"url"`
}

// CommitStats represents leaderboard stat
type CommitStats struct {
	AuthorName  string `json:"author_name"`
	CommitCount int    `json:"commit_count"`
}
