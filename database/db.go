package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
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

func ValidateUser(idToken string) (services.User, error) {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "r-j-magenta-carrot-42069"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatal(err)
		return services.User{}, err
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
		return services.User{}, err
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		log.Fatalf("error verifying ID token: %v\n", err)
		return services.User{}, err
	}

	claims := token.Claims
	user := services.User{
		Email:       claims["email"].(string),
		Uid:         claims["user_id"].(string),
		DisplayName: claims["name"].(string),
	}

	return user, nil
}

func UpdateUser(state services.State) error {
	client, ctx, err := GetClient()
	if err != nil {
		return err
	}

	_, err = client.Collection("users").Doc(state.User.Uid).Set(ctx, state)
	if err != nil {
        return err
	}

	return err
}

func UserExists(user services.User) (bool, error) {
    client, ctx, err := GetClient()
    if err != nil {
        return false, err
    }

    doc, err := client.Collection("users").Doc(user.Uid).Get(ctx)
    if err != nil {
        return false, err
    }

    return doc.Exists(), nil
}

func GetUser(user services.User) (services.User, error) {
	client, ctx, err := GetClient()
	if err != nil {
		return services.User{}, err
	}

    doc, err := client.Collection("users").Doc(user.Uid).Get(ctx)
	if err != nil {
		return services.User{}, err
	}

    var dbuser services.User 
    err = doc.DataTo(&dbuser)
	if err != nil {
		return services.User{}, err
	}

	return dbuser, err
}
