package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/danielboakye/github-repo-stats/pkg/db/postgres"
	"github.com/danielboakye/github-repo-stats/pkg/services/githubrepo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Base struct {
	mockDB sqlmock.Sqlmock
	svc    *Server
}

func setup(t *testing.T) Base {
	require := require.New(t)

	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(err)

	postgresRepo := postgres.NewRepository(db)
	logger := slog.Default()
	githubSvc := githubrepo.NewService(postgresRepo, logger, time.Now())

	apiServer := NewServer(":9000", postgresRepo, githubSvc, logger)

	return Base{
		mockDB: mockDB,
		svc:    apiServer,
	}
}

// go test -timeout 30s -run ^TestGetLeaderboard$ ./pkg/httpserver -v
func TestGetLeaderboard(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	base := setup(t)

	base.mockDB.ExpectQuery(`
				SELECT author_name, COUNT(id) as commit_count
				FROM commits
				GROUP BY author_name
				ORDER BY commit_count DESC
				LIMIT $1
			`).
		WithArgs(2).
		WillReturnRows(
			sqlmock.NewRows([]string{"author_name", "commit_count"}).
				AddRow("user1", 5).
				AddRow("user2", 3),
		).
		RowsWillBeClosed()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/leaderboard?limit=2", nil)
	r.Header.Set("Content-Type", "application/json")

	base.svc.router.ServeHTTP(w, r)

	assert.NoError(base.mockDB.ExpectationsWereMet())
	assert.Equal(http.StatusOK, w.Code)

	res := []map[string]interface{}{}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(err)
	assert.Len(res, 2)

	user1 := res[0]
	name, exits := user1["author_name"]
	assert.True(exits)
	assert.Equal(name, "user1")
	assert.Equal(user1["commit_count"], float64(5))

	user2 := res[1]
	name, exits = user2["author_name"]
	assert.True(exits)
	assert.Equal(name, "user2")
	assert.Equal(user2["commit_count"], float64(3))
}
