FROM golang:1.22.3-alpine AS build-stage
WORKDIR /app
COPY . .
RUN go mod download
EXPOSE 8080
WORKDIR /app/cmd
RUN go test -v ../test
ENTRYPOINT exec go run main.go




