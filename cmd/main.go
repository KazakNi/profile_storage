package main

import (
	"users/config"
	server "users/internal"
)

func main() {

	config.LoadConfig()
	server.Run()

}
