package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// Configuration (Loaded from .env)
var (
	GoogleClientID      string
	GoogleClientSecret  string
	RedirectURL         string
	EncryptionKeyBase64 string

	OauthConfig   *oauth2.Config
	EncryptionKey []byte // Decoded encryption key

	// Cookie Names
	oauthStateCookieName = "oauthstate"
	userIDCookieName     = "userid"
)

// handleIndex serves a simple page with a login link.
func HandleIndex(c echo.Context) error {
	// Check if the user is already logged in by checking the user ID cookie
	state, err := GetState(c)
	if err == nil {
		// User is logged in, redirect to profile
		log.Printf("User is logged in with ID: %s", state.User.Uid)
		return Render(c, http.StatusOK, views.Index(views.Page(state.Recipes), state))
	}

	// User is not logged in
	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

// HandleLogin initiates the Google OAuth 2.0 flow.
func HandleLogin(c echo.Context) error {
	// 1. Generate a random state parameter
	state, err := generateRandomString(32) // Use a sufficiently long random string
	if err != nil {
		log.Printf("Error generating state: %v", err)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Failed to generate state parameter",
		)
	}

	// 2. Store the state parameter in a temporary cookie
	stateCookie := &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/oauth2/callback",              // Restrict cookie path to the callback URL
		Expires:  time.Now().Add(5 * time.Minute), // State cookie should be short-lived
		HttpOnly: true,                            // Prevent client-side JavaScript access
		// Secure:   true, // Uncomment in production with HTTPS
		SameSite: http.SameSiteLaxMode, // Recommended for CSRF protection
	}
	c.SetCookie(stateCookie)
	log.Printf("Set state cookie: %s", state)

	// 3. Construct the Google OAuth authorization URL and redirect the user
	authURL := OauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
	) // AccessTypeOffline gets a refresh token (optional)
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// handleOAuth2Callback handles the redirect from Google after authentication.
func HandleOAuth2Callback(c echo.Context) error {
	// 1. Get the state parameter from the query string
	queryState := c.QueryParam("state")
	if queryState == "" {
		log.Println("Error: State parameter missing in callback")
		return echo.NewHTTPError(http.StatusBadRequest, "State parameter missing in callback")
	}

	// 2. Get the state cookie
	stateCookie, err := c.Cookie(oauthStateCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			log.Println("Error: State cookie not found")
			return echo.NewHTTPError(http.StatusBadRequest, "State cookie not found")
		}
		log.Printf("Error reading state cookie: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error reading state cookie")
	}

	// 3. Validate the state parameter: compare query string state with cookie state
	if queryState != stateCookie.Value {
		log.Printf(
			"Error: Invalid state parameter. Query: %s, Cookie: %s",
			queryState,
			stateCookie.Value,
		)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid state parameter")
	}
	log.Printf("State validated successfully. Query: %s, Cookie: %s", queryState, stateCookie.Value)

	// 4. Remove the temporary state cookie
	deleteStateCookie := &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "", // Clear the value
		Path:     "/oauth2/callback",
		Expires:  time.Unix(0, 0), // Set expiry to the past to delete it
		HttpOnly: true,
		// Secure:   true, // Uncomment in production with HTTPS
	}
	c.SetCookie(deleteStateCookie)
	log.Println("State cookie removed.")

	// 5. Get the authorization code from the query string
	code := c.QueryParam("code")
	if code == "" {
		log.Println("Error: Code parameter missing in callback")
		return echo.NewHTTPError(http.StatusBadRequest, "Code parameter missing in callback")
	}

	// 6. Exchange the authorization code for tokens
	token, err := OauthConfig.Exchange(c.Request().Context(), code)
	if err != nil {
		log.Printf("Error exchanging code: %v", err)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Failed to exchange code for token",
		)
	}
	log.Println("Successfully exchanged code for token.")

	// 7. Fetch user information using the access token
	// We'll use the standard OpenID Connect UserInfo endpoint
	userInfoURL := "https://www.googleapis.com/oauth2/v3/userinfo"
	client := OauthConfig.Client(c.Request().Context(), token) // Client with access token
	resp, err := client.Get(userInfoURL)
	if err != nil {
		log.Printf("Failed to fetch user info: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch user info")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf(
			"Error fetching user info: status %d, body: %s",
			resp.StatusCode,
			string(bodyBytes),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Sprintf("Failed to fetch user info: status %d", resp.StatusCode),
		)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Error decoding user info: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to decode user info")
	}

	// Extract the unique user ID (the 'sub' claim)
	userID, ok := userInfo["sub"].(string)
	if !ok || userID == "" {
		log.Printf("Error: 'sub' claim missing or not a string in user info: %+v", userInfo)
		return echo.NewHTTPError(http.StatusInternalServerError, "User ID not found in user info")
	}
	// Extract the unique user name (the 'name' claim)
	userName, ok := userInfo["name"].(string)
	if !ok || userName == "" {
		log.Printf("Error: 'name' claim missing or not a string in user info: %+v", userInfo)
		return echo.NewHTTPError(http.StatusInternalServerError, "User name not found in user info")
	}
	// Extract the unique user email (the 'email' claim)
	userEmail, ok := userInfo["email"].(string)
	if !ok || userEmail == "" {
		log.Printf("Error: 'email' claim missing or not a string in user info: %+v", userEmail)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"User email not found in user info",
		)
	}
	log.Printf("Successfully fetched user info for user ID: %s", userID)

	dbUser := services.User{
		Email:       userEmail,
		Uid:         userID,
		DisplayName: userName,
	}

	_, err = GetState(c)
	if err != nil {
		log.Printf(
			"User %s with email %s does not exist in database... adding",
			dbUser.DisplayName,
			dbUser.Email,
		)

		state := services.State{
			User:    dbUser,
			Recipes: []services.Recipe{},
		}
		err := UpdateState(state)
		if err != nil {
			log.Print("Error: Unable to update state")
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"Error: Unable to update state",
			)
		}
	}

	// 8. Encrypt the user ID
	encryptedUserID, err := encryptUserID(userID, EncryptionKey)
	if err != nil {
		log.Printf("Error encrypting user ID: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to encrypt user ID")
	}
	log.Println("User ID encrypted successfully.")

	// 9. Store the encrypted user ID in a new, persistent cookie
	userIDCookie := &http.Cookie{
		Name: userIDCookieName,
		Value: base64.URLEncoding.EncodeToString(
			encryptedUserID,
		), // Base64 encode for cookie safety
		Path:     "/",                                // Accessible across the site
		Expires:  time.Now().Add(7 * 24 * time.Hour), // Example: Valid for 7 days
		HttpOnly: true,                               // Prevent client-side JavaScript access
		// Secure:   true, // Uncomment in production with HTTPS
		SameSite: http.SameSiteLaxMode, // Recommended
	}
	c.SetCookie(userIDCookie)
	log.Println("Encrypted user ID cookie set.")

	// 10. Redirect the user to a protected page (e.g., profile)
	return c.Redirect(http.StatusFound, "/")
}

// handleProfile serves a protected page showing the user ID from the cookie.
// func handleProfile(c echo.Context) error {
// 	userID, err := getUserIDFromCookie(c)
// 	if err != nil {
// 		// Cookie not found or decryption failed, redirect to login
// 		log.Printf("Access denied: %v", err)
// 		return c.Redirect(http.StatusSeeOther, "/") // Redirect to index which will offer login
// 	}
//
// 	// User is authenticated, display their ID
// 	return c.HTML(
// 		http.StatusOK,
// 		fmt.Sprintf("<h1>Welcome, User %s!</h1><p><a href=\"/logout\">Logout</a></p>", userID),
// 	)
// }

// handleLogout removes the user ID cookie.
// func handleLogout(c echo.Context) error {
// 	// Delete the user ID cookie
// 	deleteUserCookie := &http.Cookie{
// 		Name:     userIDCookieName,
// 		Value:    "",              // Clear the value
// 		Path:     "/",             // Must match the path the cookie was set with
// 		Expires:  time.Unix(0, 0), // Set expiry to the past
// 		HttpOnly: true,
// 		// Secure:   true, // Uncomment in production with HTTPS
// 	}
// 	c.SetCookie(deleteUserCookie)
//
// 	log.Println("User ID cookie removed. User logged out.")
//
// 	// Redirect to the index page
// 	return c.Redirect(http.StatusSeeOther, "/")
// }

// --- Helper Functions (mostly unchanged, adapted getUserIDFromCookie) ---

// generateRandomString generates a URL-safe random string.
func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// encryptUserID encrypts the user ID using AES-GCM.
// Returns the encrypted data (nonce + ciphertext + tag) concatenated.
func encryptUserID(userID string, key []byte) ([]byte, error) {
	plaintext := []byte(userID)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("could not generate nonce: %w", err)
	}

	// Seal encrypts and authenticates plaintext.
	// It appends the tag to the ciphertext.
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Concatenate nonce and ciphertext for storage
	encryptedData := append(nonce, ciphertext...)

	return encryptedData, nil
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

// getUserIDFromCookie reads, decodes, and decrypts the user ID cookie using Echo context.
func getUserIDFromCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie(userIDCookieName)
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
