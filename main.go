package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/danielboakye/github-repo-stats/pkg/db/postgres"
	"github.com/danielboakye/github-repo-stats/pkg/httpserver"
	"github.com/danielboakye/github-repo-stats/pkg/services/githubrepo"
)

var (
	httpHost string
	httpPort string

	commitSinceDateString string
	defaultSinceDate      string
)

func init() {
	// default since date
	durationOneYear := time.Hour * 24 * 365
	defaultSinceDate = time.Now().Add(-durationOneYear).Format(githubrepo.ISODateFormat)

	flag.StringVar(&httpHost, "host", "localhost", "The host address where the application will run")
	flag.StringVar(&httpPort, "port", "9000", "The http server port")
	flag.StringVar(&commitSinceDateString, "since", defaultSinceDate, "date to start pulling commits from")
}

func main() {
	flag.Parse()

	if commitSinceDateString == "" {
		commitSinceDateString = defaultSinceDate
	}
	commitSinceDate, err := time.Parse(githubrepo.ISODateFormat, commitSinceDateString)
	if err != nil {
		log.Fatal("failed to parse 'since' flag into iso date format")
	}

	conn, err := postgres.NewConnection(os.Getenv(postgres.PostgresURLEnvVar))
	if err != nil {
		log.Fatal("failed to start postgres db: ", err)
	}

	postgresRepo := postgres.NewRepository(conn)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	githubSvc := githubrepo.NewService(postgresRepo, logger, commitSinceDate)
	if err := githubSvc.Start(context.Background()); err != nil {
		log.Fatal("failed to start background service: ", err)
	}

	logger.Info("starting-watcher",
		slog.String("since", commitSinceDate.Format(githubrepo.ISODateFormat)),
	)

	addr := fmt.Sprintf("%s:%s", httpHost, httpPort)
	apiServer := httpserver.NewServer(addr, postgresRepo, githubSvc, logger)
	if err := apiServer.Start(); err != nil {
		log.Fatal("failed to start http server on: ", addr)
	}
}
