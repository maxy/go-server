package main

import (
	"net/http"
	"strconv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mattn/go-jsonpointer"
	"os"
	"os/signal"
	"context"
	"time"
	"io/ioutil"
	"encoding/json"
)

const configFilePath = "config/config.json"

type (
	user struct {
		ID		int		`json:"id"`
		Name 	string 	`json:"name"`
	}
)

type Config struct {
	filePath	string
	raw			interface{}
}

var (
	users 	= map[int]*user{}
	seq 	= 1
)

func (config *Config) isExistConfig() bool {
	_, err := os.Stat(config.filePath)
	return err == nil
}

func (config *Config) load() bool {
	if config.isExistConfig() {
		raw, err := ioutil.ReadFile(config.filePath)
		if err == nil {
			json.Unmarshal([]byte(raw), &config.raw)
		} else {
			return false
		}
	}
	return true
}

func (config *Config) getString(pointer string) string {
	value, err := jsonpointer.Get(config.raw, pointer)
	if err != nil {
		panic(err)
	} else {
		return value.(string)
	}
}

func (config *Config) getInt(pointer string) int64 {
	value, err := jsonpointer.Get(config.raw, pointer)
	if err != nil {
		panic(err)
	} else {
		return value.(int64)
	}
}

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
	// Load config
	config := Config{filePath: configFilePath}
	config.load()

	e := echo.New()

	// Middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: config.getString("/log/format"),
	}))
	e.Use(middleware.Recover())

	e.Logger.Debugf("%v", config)

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