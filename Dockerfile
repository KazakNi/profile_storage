FROM golang:1.22
WORKDIR /app
COPY . .
RUN go mod download
EXPOSE 8080
WORKDIR /app/cmd
ENTRYPOINT exec go run main.go


