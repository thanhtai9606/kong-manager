# Kong Manager — Vue SPA + Go BFF in one image.
# Build: docker build -t kong-manager:latest .
# Run:  see docker-compose.yml or pass env (DATABASE_*, KONG_ADMIN_URL, JWT_SECRET, …).

# --- Frontend (Vite → dist/) ---
FROM node:22-bookworm AS fe
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@9.9.0 --activate

COPY package.json pnpm-lock.yaml ./
ENV CI=true
RUN pnpm install --frozen-lockfile --ignore-scripts

COPY index.html vite.config.ts tsconfig.json eslint.config.mjs .stylelintrc.cjs ./
COPY public ./public
COPY src ./src
RUN pnpm run build

# --- Backend ---
FROM golang:1.25-bookworm AS be
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /kong-manager ./cmd/kong-manager

# --- Runtime (static non-root) ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=be /kong-manager /app/kong-manager
COPY --from=fe /app/dist /app/dist

ENV STATIC_DIR=/app/dist
ENV HTTP_ADDR=:8080

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/kong-manager"]
