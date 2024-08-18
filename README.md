# Project Setup and Usage

This project requires [Docker](https://www.docker.com/) and [Make](https://www.gnu.org/software/make/) to manage the setup, starting the application, and interacting with it through specific commands.

## Prerequisites

- **[Docker](https://www.docker.com/)**: Ensure that Docker is installed and running on your machine.
- **[Make](https://www.gnu.org/software/make/)**: Ensure that `make` is installed on your system.

## Setup

### 1. Setting Up the Database

To set up the PostgreSQL database using Docker, run the following command:

```bash
make setup/db
```

### 2. Starting the Application

To start the application, you can use the following command:

```bash
make start
```

By default, the application will start on port 9000. If you want to use a different port, you can specify it like this:

```bash
make start PORT=8080
```

### 3. Get the top N commit authors by commit counts from the database

```bash
make get-leaderboard
```

This defaults to

```bash
make get-leaderboard LIMIT=5
```

it runs `curl -s -X GET http://localhost:9000/v1/leaderboard?limit=5 "Accept: application/json"`

You can specify number of authors (N) in leaderboard by like this using LIMIT.

```bash
make get-leaderboard LIMIT=10
```

### 4. Retrieve commits of a repository by repository name from the database

```bash
make get-commits
```

This defaults to

```bash
make get-commits REPO=chromium/chromium
```

it runs `curl -s -X GET "http://localhost:9000/v1/commits?repoName=chromium/chromium&limit=5" -H "Accept: application/json"`

If you want to get the commits for a different repository, you can specify the repository like this:

```bash
make get-commits REPO=mozilla/gecko-dev
```

### 5. Reset the collection to start from a point in time

Reset the database by running

```bash
make setup/db
```

add date in ISO 8601 format: YYYY-MM-DDTHH:MM:SSZ to start command

```bash
make start SINCE="2023-08-19T04:28:03Z"
```

#### NOTES:

- Please make sure you have `make start` running in a different terminal before running `make get-commits` or `make get-leaderboard` in a second terminal
- running `make get-commits REPO=mozilla/gecko-dev` will start tracking `<mozilla/gecko-dev>` if it is not tracking
- To track a new repo run `make get-commits REPO=owner/name`

#### TROUBLESHOOTING:

- `make: [*] Error 7` - Please make sure you are using the same PORT value in the flag for both the `make start` and `make get-commit` and `make get-leaderboard`
  Example
  - `make start PORT=8080`
  - `make get-commits REPO=chromium/chromium LIMIT=5 PORT=8080`
  - `make get-leaderboard LIMIT=5 PORT=8080`
