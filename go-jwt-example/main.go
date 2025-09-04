package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// IMPORTANT: In a real application, this key should be loaded from a secure
// configuration and should be much more complex.
var jwtKey = []byte("my_secret_key")

// Claims struct defines the structure of the JWT payload.
// We can include standard claims and our own custom claims.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	fmt.Println("--- JWT Creation and Validation Example ---")

	// --- Step 1: Create a new token ---
	username := "john.doe"
	tokenString, err := createToken(username)
	if err != nil {
		fmt.Println("Error creating token:", err)
		return
	}
	fmt.Printf("\nGenerated JWT for user '%s':\n%s\n", username, tokenString)

	// --- Step 2: Validate the token ---
	// Let's simulate a client sending this token back to the server.
	fmt.Println("\n--- Validating the token --- ")
	validateToken(tokenString)

	// --- Step 3: Try to validate a tampered token ---
	fmt.Println("\n--- Validating a tampered token --- ")
	tamperedToken := tokenString + "X"
	validateToken(tamperedToken)

	// --- Step 4: Try to validate an expired token ---
	fmt.Println("\n--- Validating an expired token --- ")
	expiredToken, _ := createExpiredToken(username)
	validateToken(expiredToken)
}

// createToken generates a new JWT for a given username.
func createToken(username string) (string, error) {
	// Set expiration time for the token (e.g., 5 minutes).
	expirationTime := time.Now().Add(5 * time.Minute)

	// Create the JWT claims, which includes the username and standard claims.
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Create a new token object, specifying the signing method and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key to get the complete, signed token string.
	tokenString, err := token.SignedString(jwtKey)

	return tokenString, err
}

// validateToken parses and validates a token string.
func validateToken(tokenString string) {
	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in the anonymous function callback.
	// This is the recommended way to handle verification to prevent
	// attacks where the signing method is changed.
	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the signing method is what you expect.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			fmt.Println("Validation Error: Invalid signature")
			return
		}
		// Check for other errors, like expiry.
		fmt.Println("Validation Error:", err)
		return
	}

	if !tkn.Valid {
		fmt.Println("Validation Error: Token is not valid")
		return
	}

	fmt.Printf("Token is valid! Welcome, %s!\n", claims.Username)
}

// createExpiredToken is a helper to demonstrate expiry validation.
func createExpiredToken(username string) (string, error) {
	expirationTime := time.Now().Add(-5 * time.Minute) // Expired 5 minutes ago
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
