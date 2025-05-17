package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/mwdev22/TCP-bomberman/pkg/config"
)

func main() {

	cfg := config.NewClientConfig()

	conn, err := net.Dial("tcp", cfg.Host+":"+cfg.Port)
	if err != nil {
		log.Fatalf("unable to connect to server: %v\n", err)
	}
	defer conn.Close()

	fmt.Println("connected to Bomberman server!")

	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Disconnected from server.")
				os.Exit(0)
			}
			fmt.Print("Server: " + msg)
		}
	}()

	console := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if console.Scan() {
			text := console.Text()
			_, err := fmt.Fprintf(conn, "%s\n", text)
			if err != nil {
				log.Printf("send error: %v\n", err)
				return
			}
		}
	}
}
