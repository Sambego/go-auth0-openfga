package main

import (
	"errors"
	"slices"

	"github.com/google/uuid"
)

type Picture struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type Store struct {
	Pictures []Picture
}

func (store *Store) getPictures() *[]Picture {
	return &store.Pictures
}

func (store *Store) getPicture(id string) (*Picture, error) {
	// Find the picture index in our store
	idx := slices.IndexFunc(store.Pictures, func(picture Picture) bool { return picture.Id == id })

	if idx < 0 {
		return nil, errors.New("Picture not found")
	}

	// Return the picture
	return &store.Pictures[idx], nil
}

func (store *Store) CreatePicture(picture Picture) (*Picture, error) {
	// Create a new UUID
	id := uuid.NewString()
	picture.Id = id

	// Add the picture to our datastore
	store.Pictures = append(store.Pictures, picture)

	return &picture, nil
}

func (store *Store) DeletePicture(id string) error {
	// Find the picture index in our store
	idx := slices.IndexFunc(store.Pictures, func(picture Picture) bool { return picture.Id == id })

	if idx < 0 {
		return errors.New("Picture not found")
	}

	// Delete the picture
	store.Pictures = append(store.Pictures[:idx], store.Pictures[idx+1:]...)
	return nil
}
