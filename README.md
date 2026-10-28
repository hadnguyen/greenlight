# Running the project

Requirements: Go 1.24.2 or later and Docker Compose. Run the commands below from the project root.

1. Create or update the `.env` file:

   ```env
   GREENLIGHT_DB_DSN=postgres://admin:admin@localhost:5532/greenlight?sslmode=disable
   ```

2. Start PostgreSQL:

   ```sh
   docker compose up -d
   ```

   Wait until the database is ready (the command below returns `accepting connections`):

   ```sh
   docker compose exec postgres pg_isready -U admin -d greenlight
   ```

3. Apply the database migrations using the [migration instructions](#database-migrations) below.

4. Download dependencies and run the API:

   ```sh
   go mod download
   go run ./cmd/api
   ```

5. Check the API from another terminal:

   ```sh
   curl http://localhost:4000/v1/healthcheck
   ```

## Database migrations

Run these commands from the project root with PostgreSQL running. Replace `<DATABASE_DSN>` with your PostgreSQL connection string (the `GREENLIGHT_DB_DSN` value in `.env`).

```sh
# Install the CLI (macOS)
brew install golang-migrate

# Create a new pair of .up.sql and .down.sql files
# Replace add_example_column with a descriptive migration name
migrate create -seq -ext=.sql -dir=./migrations add_example_column

# Apply all pending migrations
migrate -path=./migrations -database='<DATABASE_DSN>' up

# Check the current migration version
migrate -path=./migrations -database='<DATABASE_DSN>' version

# Roll back the latest migration
migrate -path=./migrations -database='<DATABASE_DSN>' down 1

# Move to a specific version (example: version 1)
migrate -path=./migrations -database='<DATABASE_DSN>' goto 1
```

Fill in the generated SQL files before running `up`. Rollbacks may delete data.

For a dirty migration, repair the schema first, then set the matching version (this does not run SQL):

```sh
migrate -path=./migrations -database='<DATABASE_DSN>' force 1
```

## Run with live reload

Install Air once (requires Go 1.25 or later):

```sh
go install github.com/air-verse/air@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

After completing the database setup above, run this from the project root instead of `go run ./cmd/api`:

```sh
air
```

## API documentation

With the API running, open http://localhost:4000/swagger/index.html to browse and try the endpoints.

Air automatically regenerates the documentation before each build when a watched Go file changes. When running without Air, regenerate the documentation manually from the project root after changing API annotations:

```sh
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go --parseInternal
```

Keep the generated files in `docs/` alongside the code changes. Air excludes `docs/` from file watching to prevent reload loops.

## Stop the project

Stop the API with `Ctrl+C`, then stop the database:

```sh
docker compose down
```
