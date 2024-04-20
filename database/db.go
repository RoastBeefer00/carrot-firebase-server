package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

func GetClient() (*firestore.Client, context.Context, error) {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "r-j-magenta-carrot-42069"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	client, err := app.Firestore(ctx)
	return client, ctx, err
}

func ValidateUser(idToken string) {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "r-j-magenta-carrot-42069"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		log.Fatalf("error verifying ID token: %v\n", err)
	}

	log.Printf("Verified ID token: %v\n", token)
    claims := token.Claims
	log.Printf("Verified ID claims: %v\n", claims)
	log.Printf("Email: %v\n", claims["email"])
	log.Printf("Name: %v\n", claims["name"])
	log.Printf("ID: %v\n", claims["user_id"])
}
