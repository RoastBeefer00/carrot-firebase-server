package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/labstack/echo/v4"
)

func GetToken(c echo.Context) ( string, error ) {
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
    new_state := services.State{
        User: services.User{},
        Recipes: []services.Recipe{},
    }
    token, err := GetToken(c)
    if err != nil {
        return new_state, err
    }

    user, err := ValidateUser(token)
    if err != nil {
        return new_state, err
    }
    new_state.User = user

	client, ctx, err := GetClient()
	if err != nil {
		return new_state, err
	}

    doc, err := client.Collection("users").Doc(user.Uid).Get(ctx)
	if err != nil {
        if doc.Exists() == false {
            log.Printf("User %s with email %s does not exist in database... adding", user.DisplayName, user.Email)
            err := UpdateState(new_state)
            if err != nil {
                return new_state, err
            }
        }
		return new_state, err
	}


    var dbuser services.State 
    err = doc.DataTo(&dbuser)
	if err != nil {
		return new_state, err
	}

	return dbuser, err
}
