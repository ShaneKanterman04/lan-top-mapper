FROM node:18-alpine AS web-build
WORKDIR /app/web
COPY web/package.json ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.24 AS go-build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . ./
COPY --from=web-build /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/lan-topology-mapper ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends iputils-ping iproute2 && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /app/data
COPY --from=go-build /out/lan-topology-mapper /app/lan-topology-mapper
COPY --from=go-build /app/migrations /app/migrations
COPY --from=go-build /app/web/dist /app/web/dist
EXPOSE 8080
ENV APP_ADDR=:8080
CMD ["/app/lan-topology-mapper"]
