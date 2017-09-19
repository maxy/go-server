package main

import (
	"net/http"
	"strconv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
	"os/signal"
	"context"
	"time"
)

type (
	user struct {
		ID		int		`json:"id"`
		Name 	string 	`json:"name"`
	}
)

var (
	users 	= map[int]*user{}
	seq 	= 1
)

// --------
// Handlers
// --------

func root(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}

func createUser(c echo.Context) error {
	u := &user{
		ID: seq,
	}
	if err := c.Bind(u); err != nil {
		return err
	}
	users[u.ID] = u
	seq ++
	return c.JSON(http.StatusCreated, u)
}

func getUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if users[id] != nil {
		return c.JSON(http.StatusOK, users[id])
	} else {
		return c.NoContent(http.StatusNotFound)
	}
}

func updateUser(c echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	users[id].Name = u.Name
	return c.JSON(http.StatusOK, users[id])
}

func deleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	delete(users, id)
	return c.NoContent(http.StatusNoContent)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", root)
	e.POST("/users", createUser)
	e.GET("/users/:id", getUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	go func() {
		if err := e.Start(":1323"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}