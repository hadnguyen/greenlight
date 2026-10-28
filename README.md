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

3. Create the table and constraints (run only once for a new database):

   ```sh
   docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U admin -d greenlight < migrations/000001_create_movies_table.up.sql
   docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U admin -d greenlight < migrations/000002_add_movies_check_constraints.up.sql
   ```

4. Download dependencies and run the API:

   ```sh
   go mod download
   go run ./cmd/api
   ```

5. Check the API from another terminal:

   ```sh
   curl http://localhost:4000/v1/healthcheck
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

## Stop the project

Stop the API with `Ctrl+C`, then stop the database:

```sh
docker compose down
```
