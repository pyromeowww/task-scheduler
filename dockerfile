FROM golang:1.26.1-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o task-scheduler .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/task-scheduler .
COPY web ./web
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/data/scheduler.db
CMD ["./task-scheduler"]
