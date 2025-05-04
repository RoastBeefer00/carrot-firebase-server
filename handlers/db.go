package handlers

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/labstack/echo/v4"
)

func GetToken(c echo.Context) (string, error) {
	cook, err := c.Cookie("token")
	if err != nil {
		return "", err
	}
	token := cook.Value

	return token, nil
}

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

func UpdateState(state services.State) error {
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

func GetState(c echo.Context) (services.State, error) {
	var state services.State
	uid, err := getUserIDFromCookie(c)
	if err != nil {
		return state, err
	}

	client, ctx, err := GetClient()
	if err != nil {
		return state, err
	}

	doc, err := client.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		if doc.Exists() == false {
			// err := UpdateState(new_state)
			// if err != nil {
			return state, err
			// }
		}
		return state, err
	}

	var dbuser services.State
	err = doc.DataTo(&dbuser)
	if err != nil {
		return state, err
	}

	return dbuser, err
}
