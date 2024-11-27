package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
)

type pictureService struct {
	Store *Store
	FGA   *FGA
}

func NewPictureService(store *Store, fga *FGA) *pictureService {
	return &pictureService{
		Store: store,
		FGA:   fga,
	}
}

func (ps *pictureService) getPicturesHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("- Getting all pictures")

	pictures := ps.Store.getPictures()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader((http.StatusOK))
	json.NewEncoder(w).Encode(pictures)
}

func (ps *pictureService) getPictureHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	log.Printf("- Getting picture: %v", id)

	// Get the authenticated user
	token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	user := token.RegisteredClaims.Subject

	// Get the custom claims
	customClaims, ok := token.CustomClaims.(*CustomClaims)

	if !ok {
		w.WriteHeader((http.StatusInternalServerError))
		return
	}

	allowed, err := ps.FGA.CheckWithRoles(fmt.Sprintf("picture:%s", id), "can_view", fmt.Sprintf("user:%s", user), customClaims.Roles)

	if err != nil || !allowed {
		log.Printf("Unauthorized", err)
		w.WriteHeader((http.StatusUnauthorized))
		w.Write([]byte(err.Error()))
		return
	}

	// Get the picture from the data store
	picture, err := ps.Store.getPicture(id)

	// Error handling
	if err != nil {
		w.WriteHeader((http.StatusInternalServerError))
		w.Write([]byte(err.Error()))
		return
	}

	// Return it all
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader((http.StatusOK))
	json.NewEncoder(w).Encode(picture)
}

func (ps *pictureService) postPictureHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("- Creating picture")

	// Get the authenticated user
	token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	user := token.RegisteredClaims.Subject

	// Parse the request body
	var parsedBody Picture
	err := json.NewDecoder(r.Body).Decode(&parsedBody)

	if err != nil {
		log.Printf("Error parsing the body", err)
		w.WriteHeader((http.StatusInternalServerError))
		w.Write([]byte(err.Error()))
		return
	}

	// Create the picture in the store
	picture, err := ps.Store.CreatePicture(parsedBody)

	if err != nil {
		log.Printf("Error getting the picture", err)
		w.WriteHeader((http.StatusInternalServerError))
		w.Write([]byte(err.Error()))
		return
	}

	// Create the ownership tupple for the current user
	// All pictures have public access by default
	// All pictures belong to the root systen
	tuples := []openfga.TupleKey{
		{
			Object:   fmt.Sprintf("picture:%s", picture.Id),
			Relation: "owner",
			User:     fmt.Sprintf("user:%s", user),
		},
		{
			Object:   fmt.Sprintf("picture:%s", picture.Id),
			Relation: "viewer",
			User:     "user:*",
		},
		{
			Object:   fmt.Sprintf("picture:%s", picture.Id),
			Relation: "system",
			User:     "system:root",
		},
	}

	// Create an FGA request
	body := client.ClientWriteRequest{
		Writes: tuples,
	}

	// Makethe FGA request
	_, err = ps.FGA.Client.Write(r.Context()).Body(body).Execute()

	if err != nil {
		log.Printf("Unauthorized", err)
		w.WriteHeader((http.StatusUnauthorized))
		w.Write([]byte(err.Error()))
		return
	}

	log.Printf("- Creating picture complete: %v", picture.Id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(picture)
}

func (ps *pictureService) deletePictureHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	log.Printf("- Deleting picture: %v", id)

	// Get the authenticated user
	token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	user := token.RegisteredClaims.Subject

	// Get the custom claims
	customClaims, ok := token.CustomClaims.(*CustomClaims)

	if !ok {
		w.WriteHeader((http.StatusInternalServerError))
		return
	}

	allowed, err := ps.FGA.CheckWithRoles(fmt.Sprintf("picture:%s", id), "can_view", fmt.Sprintf("user:%s", user), customClaims.Roles)

	if err != nil || !allowed {
		log.Printf("Unauthorized", err)
		w.WriteHeader((http.StatusUnauthorized))
		w.Write([]byte(err.Error()))
		return
	}

	err = ps.Store.DeletePicture(r.PathValue("id"))

	if err != nil {
		w.WriteHeader((http.StatusInternalServerError))
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader((http.StatusOK))
}
