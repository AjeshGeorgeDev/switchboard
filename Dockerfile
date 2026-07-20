FROM node:20-alpine AS frontend
WORKDIR /app/frontend
RUN corepack enable && corepack prepare pnpm@10 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

FROM golang:1.25-alpine AS backend
WORKDIR /app
RUN apk add --no-cache git
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend /app/backend/internal/static/dist ./internal/static/dist
RUN CGO_ENABLED=0 go build -o /server ./cmd/server
RUN CGO_ENABLED=0 go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=backend /server /server
COPY --from=backend /go/bin/migrate /usr/local/bin/migrate
COPY backend/migrations /migrations
COPY backend/rbac_model.conf /rbac_model.conf
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/docker-entrypoint.sh"]
