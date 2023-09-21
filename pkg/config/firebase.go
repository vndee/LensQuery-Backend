package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

var FirebaseApp *firebase.App
var FirebaseAuth *auth.Client

func SetupFirebase() {
	var err error

	FirebaseApp, _ = firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("[Firebase Err] Initializing app: %v\n", err)
	}

	FirebaseAuth, err = FirebaseApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("[Firebase Err] Getting Auth client: %v\n", err)
	}
}
