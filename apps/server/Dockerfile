FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app-server ./cmd/app-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/recorder-worker ./cmd/recorder-worker

FROM alpine:3.21

RUN adduser -D -H -u 10001 appuser
WORKDIR /app
COPY --from=builder /out/app-server /app/app-server
COPY --from=builder /out/recorder-worker /app/recorder-worker
USER appuser

EXPOSE 8080 8082
CMD ["/app/app-server"]
