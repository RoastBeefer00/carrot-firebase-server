package db

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/labstack/echo/v4"
)

var (
	EncryptionKey    []byte // Decoded encryption key
	UserIDCookieName = "userid"
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

func UpdateState(state services.State, c echo.Context) error {
	ctx := c.Request().Context()
	client := c.Get("db").(*firestore.Client)

	_, err := client.Collection("users").Doc(state.User.Uid).Set(ctx, state)
	if err != nil {
		return err
	}

	return err
}

func GetStateFromId(uid string, c echo.Context) (services.State, error) {
	var state services.State

	ctx := c.Request().Context()
	client := c.Get("db").(*firestore.Client)

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

func GetState(c echo.Context) (services.State, error) {
	client := c.Get("db").(*firestore.Client)
	var state services.State
	uid, err := getUserIDFromCookie(c)
	if err != nil {
		return state, err
	}

	doc, err := client.Collection("users").Doc(uid).Get(c.Request().Context())
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

// getUserIDFromCookie reads, decodes, and decrypts the user ID cookie using Echo context.
func getUserIDFromCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie(UserIDCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return "", fmt.Errorf("user ID cookie not found")
		}
		return "", fmt.Errorf("error reading user ID cookie: %w", err)
	}

	// Decode the Base64 encoded cookie value
	encryptedData, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("failed to decode user ID cookie: %w", err)
	}

	// Decrypt the data
	userID, err := decryptUserID(encryptedData, EncryptionKey)
	if err != nil {
		// Decryption failed, cookie is invalid or tampered
		// In a real app, you might want to clear the invalid cookie here.
		return "", fmt.Errorf("failed to decrypt user ID cookie: %w", err)
	}

	return userID, nil
}

// decryptUserID decrypts the encrypted data using AES-GCM.
func decryptUserID(encryptedData []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("could not create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("could not create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertextWithTag := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Open decrypts and authenticates ciphertext.
	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		// This error likely means the data was tampered with or the key is wrong
		return "", fmt.Errorf("could not decrypt or authenticate data: %w", err)
	}

	return string(plaintext), nil
}
