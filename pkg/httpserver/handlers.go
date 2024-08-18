package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/danielboakye/github-repo-stats/pkg/response"
)

const (
	repoNameQueryParam = "repoName"
	limitQueryParam    = "limit"
	offsetQueryParam   = "offset"
)

// ValidateRepoName checks if the repository name is in the correct "owner/name" format.
// for tracking in the context of dynamic owner repos requiring full repo name with owner prefix is necessary
// as repo names are only unique per owner not in the whole github
func ValidateRepoName(repoName string) error {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("repository name must be in the format 'owner/name'")
	}

	owner, name := parts[0], parts[1]
	if owner == "" || name == "" {
		return fmt.Errorf("both owner and repository name must be non-empty")
	}

	return nil
}

// GetCommits is the http handler for GetCommits in github svc
func (s *Server) GetCommits(w http.ResponseWriter, r *http.Request) {
	repoName := strings.ToLower(strings.TrimSpace(r.URL.Query().Get(repoNameQueryParam)))
	if repoName == "" {
		response.InvalidRequest(w, "repoName is missing")
		return
	}
	err := ValidateRepoName(repoName)
	if err != nil {
		response.InvalidRequest(w, err.Error())
		return
	}

	limitStr := r.URL.Query().Get(limitQueryParam)
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	offsetStr := r.URL.Query().Get(offsetQueryParam)
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	commits, err := s.githubSvc.GetCommits(r.Context(), repoName, limit, offset)
	if err != nil {
		s.logger.Error("failed-getting-commits",
			slog.String("path", "getCommits"),
			slog.String("error", err.Error()),
		)
		response.InternalError(w)
		return
	}
	if len(commits) == 0 {
		if err := response.JSON(w, http.StatusAccepted, map[string]string{
			"message": fmt.Sprintf("%s is now being tracked. Please check back later", repoName),
		}); err != nil {
			s.logger.Error("failed-encoding-json",
				slog.String("path", "getCommits"),
			)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, commits); err != nil {
		s.logger.Error("failed-encoding-json",
			slog.String("path", "getCommits"),
		)
	}
}

// GetLeaderBoard is the http handler for GetLeaderBoard in github svc
func (s *Server) GetLeaderBoard(w http.ResponseWriter, r *http.Request) {
	countStr := r.URL.Query().Get(limitQueryParam)
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		count = 5
	}

	leaderBoard, err := s.githubSvc.GetLeaderBoard(r.Context(), count)
	if err != nil {
		s.logger.Error("failed-getting-leader-board",
			slog.String("path", "getLeaderBoard"),
			slog.String("error", err.Error()),
		)
		response.InternalError(w)
		return
	}
	if len(leaderBoard) == 0 {
		if err := response.JSON(w, http.StatusAccepted, map[string]string{
			"message": "no repositories are currently being tracked",
		}); err != nil {
			s.logger.Error("failed-encoding-json",
				slog.String("path", "getLeaderBoard"),
			)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, leaderBoard); err != nil {
		s.logger.Error("failed-encoding-json",
			slog.String("path", "getLeaderBoard"),
		)
	}
}

// NotFoundHandler handles all unfamiliar routes
func (s *Server) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	if err := response.JSON(w, http.StatusNotFound, map[string]string{
		"error":   "Not Found",
		"message": "The requested resource could not be found",
	}); err != nil {
		s.logger.Error("failed-encoding-json",
			slog.String("path", "notFound"),
		)
	}
}
