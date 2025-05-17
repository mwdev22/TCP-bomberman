package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ClientConfig struct {
	Host string
	Port string
}

type ServerConfig struct {
	Addr string
	Port string
}

func NewServerConfig() *ServerConfig {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file")
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "localhost"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return &ServerConfig{
		Addr: addr,
		Port: port,
	}

}

func NewClientConfig() *ClientConfig {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file")
	}

	host := os.Getenv("SERVER_ADDR")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return &ClientConfig{
		Host: host,
		Port: port,
	}

}
