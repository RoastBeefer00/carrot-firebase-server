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

    client, err :=  app.Firestore(ctx)
    return client, ctx, err
}
