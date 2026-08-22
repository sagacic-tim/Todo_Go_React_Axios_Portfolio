# backend/Dockerfile

# ---- build stage ----
FROM golang:1.24-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

# Cache deps first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the backend source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/server

# ---- runtime stage ----
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

# App binary
COPY --from=build /out/api /app/api

# IMPORTANT: migrations must exist at runtime because you run them on startup
COPY --from=build /src/migrations /app/migrations

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
