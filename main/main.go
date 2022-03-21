package main

import (
	"github.com/marceloSantosC/go-server"
	"log"
	"net/http"
	"os"
	"strconv"
)

type AppConfig struct {
	port    int
	appName string
}

func main() {

	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		port = 8080
		log.Println("Port is not defined, starting at default port (8080)")
	}
	config := AppConfig{port, "go-books"}

	handlers := createHandlers()
	server.CreateServer(config.port, handlers)
}

func createHandlers() map[string]func(w http.ResponseWriter, r *http.Request) {

	handlers := make(map[string]func(rw http.ResponseWriter, req *http.Request))

	handlers["/books"] = booksHandler

	return handlers
}
