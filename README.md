Todo Axios API + Calendar App (Portfolio Snapshot)

This is a sanitized public snapshot of the private production repository,
published for code review by prospective employers. It is squashed to a
single commit (no production git history) and has real .env values
replaced with .env.example placeholders — everything else reflects real,
running code.

A full-stack task management application built around a calendar-grid interface where tasks are scheduled and managed by day.

The project demonstrates a modern architecture using:

  Go (Gin + GORM) backend API
  PostgreSQL database
  React + TypeScript + Vite frontend
  Docker Compose containerized deployment
  Nginx reverse proxy + TLS
  GitHub Container Registry (GHCR) for CI/CD images

The goal of the project is to explore a clean separation between:

  API service
  data persistence
  frontend UI
  infrastructure
  and to demonstrate a realistic deployment pipeline similar to what is used in production systems.

Application Overview

  The application presents a calendar grid interface where each day can contain multiple tasks.

Each task includes:

  title
  description
  due date
  state (scheduled, rescheduled, completed, dismissed)

Tasks are displayed in the calendar interface and can be edited via popup editors.

The system supports:

  creating tasks
  editing tasks
  changing task state
  deleting tasks
  retrieving tasks sorted by due date

Architecture

  Browser
    │
    │ HTTPS
    ▼
  Nginx Reverse Proxy
    │
    ├── /api → Go API (Gin)
    │
    └── / → React SPA
            │
            ▼
          Axios
            │
            ▼
          Go API
            │
            ▼
        PostgreSQL

The application is split into two services:

Backend Service

  Language: Go
  Frameworks/Libraries:
  Gin (HTTP routing)
  GORM (ORM)
  PostgreSQL driver
  Docker container runtime

Responsibilities:

  REST API
  database access
  migrations
  task lifecycle management

API endpoints:

  GET    /api/tasks
  POST   /api/tasks
  PATCH  /api/tasks/:id
  DELETE /api/tasks/:id

Frontend Service

  Framework: React + TypeScript

Tooling:

  Vite
  Axios
  modern ES modules

Responsibilities:

  render calendar grid
  display tasks for each day
  open editor popups
  communicate with API

Technology Stack

Backend

  Go
  Gin
  GORM
  PostgreSQL

Frontend

  React
  TypeScript
  Vite
  Axios

Infrastructure

  Docker
  Docker Compose
  Nginx
  Let’s Encrypt TLS
  GitHub Container Registry (GHCR)

Repository Structure

  backend/
    cmd/server/
        main.go

    internal/
        db/
        models/
        repository/
        service/
        transport/

  frontend/
    src/
        components/
        pages/
        services/
        hooks/

docker-compose.yml

The backend follows a layered architecture:

  transport  → HTTP handlers
  service    → business logic
  repository → data access
  models     → persistence models

This separation makes the codebase easier to test and maintain.

Running Locally

Requirements:

  Docker
  Docker Compose

Start the application:

  docker compose up --build

Services will start:

  API      http://localhost:8081
  Frontend http://localhost:8082

Environment Configuration

  Example .env file:

    POSTGRES_USER=todo
    POSTGRES_PASSWORD=secure_password
    POSTGRES_DB=todo_axios_api

  DB_HOST=db
  DB_PORT=5432
  DB_USER=todo
  DB_PASSWORD=secure_password
  DB_NAME=todo_axios_api

The backend constructs its database DSN from these values.

Testing

Backend tests can be run with:

  go test ./...

Tests include:

  repository layer tests
  service layer tests
  transport layer tests using Gin test context
  SQLite in-memory databases are used for fast test execution.

Deployment  

  The project is designed for containerized deployment.

Typical production stack:

  Nginx
    ↓
  Docker containers
    ├── API
    ├── Web
    └── PostgreSQL

  Images are built and published to GitHub Container Registry.

Deployment workflow:

  GitHub Push
    ↓
  CI Build
    ↓
  Publish Images (GHCR)
    ↓
  VPS pulls images
    ↓
  docker compose up -d
  Security

Production deployments include:

  HTTPS via Let’s Encrypt
  Nginx reverse proxy
  server-level Basic Auth protection for demo deployments
  containerized services

Purpose of the Project

  This project was created to demonstrate practical experience with:

Go backend architecture

  REST API design
  containerized deployments
  React frontend integration
  infrastructure automation

It serves as a working reference implementation for a modern full-stack application.

License

  MIT License