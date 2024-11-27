package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func defaultHandler(writer http.ResponseWriter, request *http.Request) {
	io.WriteString(writer, "ok\n")
}

func main() {
	log.Printf("Starting server\n")
	log.Printf("---------------\n")
	// Load the .env file
	log.Printf("- Getting .env variables")
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file", err)
	}

	// Initialize the OpenFGA client, store and model
	log.Printf("- Configuring OpenFGA")
	fgaClient, err := NewFGAClient()

	if err != nil {
		log.Fatalf("Error configuring OpenFGA", err)
	}

	// Create a new data store (in memory)
	var pictures []Picture
	store := Store{
		Pictures: pictures,
	}

	// Initialize routes
	log.Printf("- Configuring routes")
	pictureService := NewPictureService(&store, fgaClient)
	router := http.NewServeMux()
	router.HandleFunc("GET /", defaultHandler)
	router.HandleFunc("GET /pictures", pictureService.getPicturesHandler)
	router.HandleFunc("GET /pictures/{id}", pictureService.getPictureHandler)
	router.HandleFunc("POST /pictures", pictureService.postPictureHandler)
	router.HandleFunc("DELETE /pictures/{id}", pictureService.deletePictureHandler)

	// Initialize server
	log.Printf("- Configuring server")
	port := ":3001"
	server := &http.Server{
		Handler:      NewEnsureValidAccessToken(router),
		Addr:         fmt.Sprintf("localhost%s", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	// Start server
	log.Printf("- Server is running on port: %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
