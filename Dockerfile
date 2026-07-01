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

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=backend /server /server
COPY backend/migrations /migrations
COPY backend/rbac_model.conf /rbac_model.conf
WORKDIR /
EXPOSE 8080
CMD ["/server"]
