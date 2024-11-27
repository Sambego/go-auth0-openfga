package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openfga/go-sdk/client"
	. "github.com/openfga/go-sdk/client"
)

type FGA struct {
	Client *OpenFgaClient
}

func NewFGAClient() (*FGA, error) {
	fga := &FGA{}
	err := fga.Setup()
	return fga, err
}

func (fga *FGA) Setup() error {
	var err error
	fga.Client, err = NewSdkClient(&ClientConfiguration{
		ApiUrl: os.Getenv("OPENFGA_DOMAIN"),
	})

	if err != nil {
		log.Fatalf("Error creating OpenFGA client", err)
		return err
	}

	// Check if there's already a store available
	listStoreOptions := ClientListStoresOptions{}
	stores, err := fga.Client.ListStores(context.Background()).Options(listStoreOptions).Execute()

	if err != nil {
		log.Fatalf("Error getting available OpenFGA stores", err)
		return err
	}

	// If store exists, use laste one, otherwise create a new one
	if len(stores.Stores) > 0 {
		// Set the new store as the store ID for the Open FGA Client
		fgaStore := stores.Stores[len(stores.Stores)-1]
		fga.Client.SetStoreId(fgaStore.Id)
		log.Printf("-- Using existing OpenFGA Store: %v", fgaStore.Id)
	} else {
		// // Create an OpenFGA Store
		fgaStore, err := fga.Client.CreateStore(context.Background()).Body(ClientCreateStoreRequest{Name: "Picture demo"}).Execute()

		if err != nil {
			log.Fatalf("Error creating a new OpenFGA store", err)
		}

		// Set the new store as the store ID for the Open FGA Client
		fga.Client.SetStoreId(fgaStore.Id)
		log.Printf("-- New OpenFGA Store: %v", fgaStore.Id)
	}

	readModelOptions := ClientReadLatestAuthorizationModelOptions{}
	latestFgaModel, err := fga.Client.ReadLatestAuthorizationModel(context.Background()).Options(readModelOptions).Execute()

	if err != nil {
		log.Fatalf("Error getting latest OpenFGA model", err)
		return err
	}

	if latestFgaModel.AuthorizationModel != nil {
		log.Printf("-- Using existing OpenFGA authorization model: %v", latestFgaModel.AuthorizationModel.Id)
	} else {
		// Read the model.json file, containing our OpenFGA model
		jsonData, err := os.ReadFile("./model/model.json")

		if err != nil {
			log.Fatal("Error reading model.json", err)
			return err
		}

		// Convert to a ClientWriteAuthorizationModelRequest
		var body ClientWriteAuthorizationModelRequest
		if err := json.Unmarshal([]byte(jsonData), &body); err != nil {
			log.Fatal("Error marshaling model JSON", err)
		}

		fgaModel, err := fga.Client.WriteAuthorizationModel(context.Background()).
			Body(body).
			Execute()

		log.Printf("-- New OpenFGA model: %v", fgaModel.AuthorizationModelId)

		if err != nil {
			log.Fatal("Error creating new OpenFGA model", err)
			return err
		}
	}

	return nil
}

func (fga *FGA) CheckWithRoles(fgaObject string, fgaRelation string, fgaUser string, roles []string) (bool, error) {
	// Create a slice with tuples for all roles
	var contextualTuples []client.ClientTupleKey
	for _, role := range roles {
		contextualTuples = append(contextualTuples, client.ClientTupleKey{
			Object:   fgaObject,
			Relation: role,
			User:     fgaUser,
		})
	}

	// Create an FGA check request
	body := client.ClientCheckRequest{
		Object:           fgaObject,
		Relation:         fgaRelation,
		User:             fgaUser,
		ContextualTuples: contextualTuples,
	}

	// Make the FGA request
	resp, err := fga.Client.Check(context.Background()).Body(body).Execute()

	// Error Handling
	if err != nil || !resp.GetAllowed() {
		log.Printf("FGA Check with roles unauthorized")
		return false, err
	}

	return true, nil
}

func (fga *FGA) CheckWithPermissions(fgaObject string, fgaRelation string, fgaUser string, permissions []string, suffix string) (bool, error) {
	// Create a slice with tuples for all permissions
	var contextualTuples []client.ClientTupleKey
	for _, permission := range permissions {
		// Check if the permission are for out current resource, eg pictures
		if strings.HasSuffix(permission, suffix) {
			// Get the actual permission from permission string
			// transform to the format we use for relations
			// eg.: delete:pictures -> can_delete
			relationPermision := fmt.Sprintf("can_%s", strings.Split(permission, ":")[0])
			contextualTuples = append(contextualTuples, client.ClientTupleKey{
				Object:   fgaObject,
				Relation: relationPermision,
				User:     fgaUser,
			})
		}
	}

	// Create an FGA request to check of the current user can view the picture
	body := client.ClientCheckRequest{
		Object:           fgaObject,
		Relation:         fgaRelation,
		User:             fgaUser,
		ContextualTuples: contextualTuples,
	}

	// Make the FGA request
	resp, err := fga.Client.Check(context.Background()).Body(body).Execute()

	// Error Handling
	if err != nil || !resp.GetAllowed() {
		log.Printf("FGA Check with permissions unauthorized")
		return false, err
	}

	return true, nil
}

func (fga *FGA) Check(fgaObject string, fgaRelation string, fgaUser string) (bool, error) {
	// Create an FGA request to check of the current user can view the picture
	body := client.ClientCheckRequest{
		Object:   fgaObject,
		Relation: fgaRelation,
		User:     fgaUser,
	}

	// Make the FGA request
	resp, err := fga.Client.Check(context.Background()).Body(body).Execute()

	// Error Handling
	if err != nil || !resp.GetAllowed() {
		log.Printf("FGA Check unauthorized")
		return false, err
	}

	return true, nil
}
