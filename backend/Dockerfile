FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o lifesim .

FROM alpine:3.20
WORKDIR /app
RUN mkdir -p /data
COPY --from=builder /app/lifesim .
COPY prompts ./prompts
ENV DATABASE_PATH=/data/lifesim.db
EXPOSE 8080
CMD ["./lifesim"]
