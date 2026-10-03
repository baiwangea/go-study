package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-faker/faker/v4"
	"io/ioutil"
	"net/http"
	"time"
)

// Define a struct with faker tags
type User struct {
	Name  string `faker:"name"`
	Email string `faker:"email"`
}

// BasicGet demonstrates how to make a simple HTTP GET request.
func BasicGet() {
	// Make a GET request to a public API.
	resp, err := http.Get("http://127.0.0.1:1323/ping")
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Print the status code and the response body.
	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Body:", string(body))
}

// GetWithHeaders demonstrates how to add custom headers to a GET request.
func GetWithHeaders() {
	// Create a new request.
	req, err := http.NewRequest("GET", "http://127.0.0.1:1323/users/muilt", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Add custom headers.
	req.Header.Set("X-Custom-Header", "my-custom-value")
	req.Header.Set("User-Agent", "my-go-app/1.0")

	// Create a client and send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Body:", string(body))
}

// PostJSON demonstrates how to send JSON data in a POST request.
func PostJSON() {
	// Create a data structure to be marshaled into JSON.
	var u User
	err := faker.FakeData(&u)
	if err != nil {
		fmt.Println(err)
	}

	// Marshal the data into a JSON byte slice.
	jsonData, err := json.Marshal(u)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Make a POST request with the JSON data.
	resp, err := http.Post("http://127.0.0.1:1323/users", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Body:", string(body))
}

// CustomClient demonstrates how to create and use a custom HTTP client with a timeout.
func CustomClient() {
	// Create a custom client with a 10-second timeout.
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Use the custom client to make a request.
	resp, err := client.Get("http://127.0.0.1:1323/users/1") // This endpoint waits 5 seconds.
	if err != nil {
		fmt.Println("Error with custom client:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Body:", string(body))
}
