FROM node:22-alpine AS frontend-builder

WORKDIR /app/ui

RUN corepack enable

COPY ui/package.json ui/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY ui/ ./
RUN pnpm run build

FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hakoniwa cmd/serve/main.go


FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=backend-builder --chown=nonroot:nonroot /app/hakoniwa .

COPY --from=frontend-builder --chown=nonroot:nonroot /app/ui/dist ./ui/dist

USER nonroot:nonroot

EXPOSE 8080

CMD ["./hakoniwa"]
