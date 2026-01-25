# Auth Microservice (Go + Fiber)

A rigorous, production-ready authentication microservice built with **Go** and **Fiber**. This service provides secure user authentication, session management, and authorization features suitable for modern distributed systems.

## 🚀 Features

- **Authentication**:
  - User Registration & Login (Email/Password).
  - **JWT** based Access & Refresh Tokens.
  - **CSRF Protection** using Double Submit Cookie pattern.
- **Session Management**:
  - Redis-backed session storage.
  - List active sessions.
  - Remote logout (single session or all sessions).
- **Security**:
  - Rate Limiting (Redis-backed).
  - Secure Cookie handling (HTTPOnly, Secure, SameSite).
  - Password Hashing.
- **Auditing**:
  - Action logs stored in PostgreSQL.
- **Password Reset**:
  - Token-based password reset flow.
- **Containerization**:
  - Docker & Docker Compose support.

## 🛠️ Tech Stack

- **Language**: Go (Golang)
- **Framework**: [Fiber](https://gofiber.io/) (v2)
- **Database**: PostgreSQL
- **Cache/Session**: Redis
- **Driver**: pgx (PostgreSQL Driver)
- **Deployment**: Docker

## 📂 Project Structure

```
├── cmd
│   └── main.go           # Application entry point
├── internal
│   ├── config            # Configuration loader
│   ├── db                # Database connection
│   ├── handler           # HTTP Route Handlers
│   ├── middlewares       # Fiber Middlewares (Auth, Security)
│   ├── models            # Domain models & DTOs
│   ├── redis             # Redis client setup
│   ├── repositories      # Data Access Objects (DAO)
│   ├── router            # Route definitions
│   ├── server            # Server configuration
│   └── services          # Business logic
├── migrations            # SQL migrations
├── Dockerfile            # Docker build file
├── docker-compose.yml    # Docker Compose setup
└── MakeFile              # Make commands
```

## ⚙️ Configuration

The application is configured via environment variables. See `.env` for defaults.

| Variable             | Description                        | Default      |
| :------------------- | :--------------------------------- | :----------- |
| `APP_PORT`           | Port to run the server on          | `8080`       |
| `DB_URL`             | PostgreSQL connection string       | **Required** |
| `REDIS_ADDR`         | Redis address (host:port)          | **Required** |
| `REDIS_PASSWORD`     | Redis password                     | `""`         |
| `JWT_ACCESS_SECRET`  | Secret for signing Access tokens   | **Required** |
| `JWT_REFRESH_SECRET` | Secret for signing Refresh tokens  | **Required** |
| `ACCESS_TOKEN_TTL`   | Access token duration (e.g. 15m)   | `15m`        |
| `REFRESH_TOKEN_TTL`  | Refresh token duration (e.g. 720h) | `720h`       |

## 🏃 Getting Started

### Prerequisites

- **Docker** and **Docker Compose** installed.
- **Go** 1.22+ (if running locally without Docker).

### Run with Docker (Recommended)

Use the included `Makefile` for easy management:

```bash
# Start all services (App, Postgres, Redis)
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

### Run Locally

1.  Ensure PostgreSQL and Redis are running.
2.  Set up your `.env` file with correct credentials.
3.  Run the application:

```bash
go run cmd/main.go
```

## 🔌 API Endpoints

### Auth

| Method | Endpoint                       | Description                                                      |
| :----- | :----------------------------- | :--------------------------------------------------------------- |
| `POST` | `/auth/register`               | Register a new user (`email`, `password`, `role`).               |
| `POST` | `/auth/login`                  | Login user. Returns `accessToken` & sets `refresh_token` cookie. |
| `POST` | `/auth/refresh`                | Refresh access token using cookie.                               |
| `POST` | `/auth/logout`                 | Logout user (clears cookies).                                    |
| `POST` | `/auth/password-reset`         | Request password reset email.                                    |
| `POST` | `/auth/password-reset/confirm` | Confirm new password with token.                                 |

### Session Management (Protected)

| Method   | Endpoint                    | Description                                |
| :------- | :-------------------------- | :----------------------------------------- |
| `GET`    | `/auth/sessions`            | List all active sessions for current user. |
| `DELETE` | `/auth/sessions/:sessionID` | Revoke a specific session.                 |
| `DELETE` | `/auth/sessions`            | Revoke all sessions (except current).      |

### Administration

| Method | Endpoint       | Description           |
| :----- | :------------- | :-------------------- |
| `GET`  | `/auth/users`  | List all users.       |
| `GET`  | `/auth/admins` | List all admin users. |

## ⚠️ Production Readiness Assessment

**Current Status**: 🟡 **Near Production Ready**

- ✅ **Architecture**: Solid clean architecture involves separation of concerns (Handlers, Services, Repositories).
- ✅ **Security**: Implements standard security practices (JWT, CSRF, Hashing).
- ✅ **Infrastructure**: Dockerized and ready for deployment.
- ❌ **Testing**: **Major Gap**. No unit or integration tests found. `*_test.go` files are missing.
- ❌ **CI/CD**: No automated build/test pipelines configured.

**Recommendation**: Before deploying to a production environment, complete the **Testing** suite to ensure reliability and regression safety.
