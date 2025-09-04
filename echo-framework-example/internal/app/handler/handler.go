package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"math/rand"
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

func GetMuiltUser(c echo.Context) error {
	user := PickUsers(users, 5)
	//if !ok {
	//	return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	//}

	return c.JSON(http.StatusOK, user)
}

// 随机取 n 个 *User，不保证顺序。
func PickUsers(m map[int]*User, n int) []*User {
	if n <= 0 || len(m) == 0 {
		return nil
	}
	if n > len(m) {
		n = len(m)
	}

	// 1. 拷出所有 key
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// 2. 洗牌
	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	// 3. 按前 n 个 key 取值
	res := make([]*User, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, m[keys[i]])
	}
	return res
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
