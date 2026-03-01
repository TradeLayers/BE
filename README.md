# TradeLayers Backend

**Related repos:** [Frontend (FE)](https://github.com/TradeLayers/FE) | [Utilities](https://github.com/TradeLayers/Utilities)

A stock trading simulation app where users can practice trading with real-time market prices without financial risk. Uses the [Finnhub API](https://finnhub.io/) for live stock data. Built with Go and Gin.

## Prerequisites

- Go 1.21+
- PostgreSQL 17
- [golang-migrate](https://github.com/golang-migrate/migrate) (for migrations)

## Setup (to run locally without Docker)

### 1. Install PostgreSQL

<details>
<summary><strong>macOS</strong></summary>

```bash
brew install postgresql@17
```
</details>

<details>
<summary><strong>Linux (Debian/Ubuntu)</strong></summary>

```bash
sudo apt update
sudo apt install postgresql-17
```
</details>

<details>
<summary><strong>Linux (Fedora/RHEL)</strong></summary>

```bash
sudo dnf install postgresql17-server postgresql17
sudo postgresql-setup --initdb
```
</details>

<details>
<summary><strong>Windows</strong></summary>

Download the installer from https://www.postgresql.org/download/windows/, or use a package manager:

```powershell
# Chocolatey
choco install postgresql17

# Scoop
scoop install postgresql
```
</details>

### 2. Add PostgreSQL to PATH

<details>
<summary><strong>macOS (Homebrew keg-only formula)</strong></summary>

Homebrew installs PostgreSQL 17 as keg-only, so you need to add it to your PATH manually:

```bash
echo 'export PATH="/opt/homebrew/opt/postgresql@17/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```
</details>

<details>
<summary><strong>Linux</strong></summary>

Usually added to PATH automatically. Verify with:

```bash
which psql
```

If not found, add the install location (e.g. `/usr/lib/postgresql/17/bin`) to your `~/.bashrc` or `~/.zshrc`.
</details>

<details>
<summary><strong>Windows</strong></summary>

The installer normally adds PostgreSQL to PATH. If not, add `C:\Program Files\PostgreSQL\17\bin` to your system `PATH` environment variable.
</details>

### 3. Start PostgreSQL

<details>
<summary><strong>macOS</strong></summary>

```bash
brew services start postgresql@17
```
</details>

<details>
<summary><strong>Linux</strong></summary>

```bash
sudo systemctl start postgresql
sudo systemctl enable postgresql   # start on boot
```
</details>

<details>
<summary><strong>Windows</strong></summary>

The PostgreSQL service usually starts automatically after installation. If not, start it from **Services** (`services.msc`) or run:

```powershell
net start postgresql-x64-17
```
</details>

### 4. Install golang-migrate

<details>
<summary><strong>macOS</strong></summary>

```bash
brew install golang-migrate
```
</details>

<details>
<summary><strong>Go install (cross-platform)</strong></summary>

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your PATH.
</details>

<details>
<summary><strong>Windows</strong></summary>

```powershell
# Scoop
scoop install migrate

# Chocolatey
choco install golang-migrate
```
</details>

### 5. Copy environment config

```bash
cp .env.example .env
```

Edit `.env` if you need to change the database user, password, or port.

### 6. Create the database user and database

<details>
<summary><strong>macOS</strong></summary>

Homebrew creates a role matching your system username, not `postgres`. Create it:

```bash
createuser -s postgres
psql -U postgres -c "ALTER USER postgres PASSWORD 'postgres';"
createdb -U postgres tradelayers
```
</details>

<details>
<summary><strong>Linux</strong></summary>

The `postgres` role exists by default. Set its password and create the database:

```bash
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'postgres';"
sudo -u postgres createdb tradelayers
```
</details>

<details>
<summary><strong>Windows</strong></summary>

The installer prompts you to set the `postgres` password during setup. Then create the database:

```powershell
createdb -U postgres tradelayers
```
</details>

### 7. Run migrations

The Makefile automatically loads your `.env` file and builds `DATABASE_URL` from the DB variables:

```bash
make migrate-up
```

### 8. Start the server

```bash
make run
```

The server starts on `http://localhost:5000`.

## Running with Docker

The full-stack Docker setup lives in the [Utilities repo](https://github.com/TradeLayers/Utilities). It runs PostgreSQL, the backend, and the frontend together using `docker-compose`.

**Prerequisites:** Docker and Docker Compose installed, and the [Utilities](https://github.com/TradeLayers/Utilities) repo cloned as a sibling directory (`../Utilities/`).

```bash
make docker-up     # Build and start all services (detached)
make docker-down   # Stop and remove all containers
make docker-logs   # Follow backend logs
```

These commands run `docker-compose` from the Utilities repo. Under the hood:
- `docker-compose up --build -d` — builds images from `../BE` and `../FE`, starts postgres + backend + frontend
- `docker-compose down` — stops everything

Once running:
- Backend: http://localhost:5000
- Frontend: http://localhost:3000

## Make Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the server |
| `make build` | Build binary to `bin/api` |
| `make test` | Run tests |
| `make migrate-up` | Apply migrations (requires `DATABASE_URL`) |
| `make migrate-down` | Rollback migrations (requires `DATABASE_URL`) |
| `make migrate-create name=<name>` | Create new migration |
| `make docker-up` | Build and start all services via docker-compose |
| `make docker-down` | Stop and remove all containers |
| `make docker-logs` | Follow backend container logs |

## API

- `GET /api/v1/health` — Health check

## Project Structure

```
cmd/api/          Entry point
internal/
  config/         Environment configuration
  database/       Database connection
  handler/        HTTP handlers
  middleware/     Middleware (logging)
  model/          GORM models
  repository/     Data access layer
  router/         Route definitions
  server/         HTTP server with graceful shutdown
  service/        Business logic
migrations/       SQL migrations
```
