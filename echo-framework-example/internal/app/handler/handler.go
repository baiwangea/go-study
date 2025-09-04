package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// User represents a user model with validation tags.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// In-memory "database"
var (
	users = map[int]*User{}
	seq   = 1
)

// Ping handles the /ping health check route.
func Ping(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "pong"})
}

// GetUser handles retrieving a user by their ID from the path.
func GetUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	user, ok := users[id]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user)
}

// CreateUser handles the creation of a new user.
func CreateUser(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if err := c.Validate(u); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Validation failed", "details": err.Error()})
	}

	u.ID = seq
	users[u.ID] = u
	seq++

	return c.JSON(http.StatusCreated, u)
}

// UpdateUser handles updating an existing user.
func UpdateUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	user, ok := users[id]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	u := new(User)
	if err := c.Bind(u); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// We only update Name and Email, not the ID.
	user.Name = u.Name
	user.Email = u.Email

	return c.JSON(http.StatusOK, user)
}

// DeleteUser handles deleting a user.
func DeleteUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if _, ok := users[id]; !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	delete(users, id)

	return c.NoContent(http.StatusNoContent)
}
