FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o doesthiswork .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/doesthiswork .
EXPOSE 8080
CMD ["./doesthiswork", "serve", "--http=0.0.0.0:8080", "--dir=/data/pb_data"]
