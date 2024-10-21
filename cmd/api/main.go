package main

import (
	"log"
)

// main is the entry point for the application.
func main() {
	// Initialize the auth service
	server, err := InitializeAuthService()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Run the server
	if err = server.Run(":8082"); err != nil {
		log.Fatal(err.Error())
	}
}
