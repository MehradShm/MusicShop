package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type User struct {
	Username string `json:"username" validate:"required"`
	ID       int    `json:"id"`
	Email    string `json:"email" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
}

type UserDB struct {
	mu     sync.RWMutex
	data   map[int]User
	nextID int
}

func NewUserDB() *UserDB {
	return &UserDB{
		data:   make(map[int]User),
		nextID: 1,
	}
}

func (s *UserDB) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]User, 0, len(s.data))
	for _, b := range s.data {
		out = append(out, b)
	}
	return out
}

func (s *UserDB) Get(id int) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.data[id]
	if !ok {
		return User{}, errors.New("not found")
	}
	return b, nil
}

func (s *UserDB) Create(b User) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	b.ID = s.nextID
	s.nextID++
	s.data[b.ID] = b
	return b
}

func (s *UserDB) Update(id int, b User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return User{}, errors.New("not found")
	}
	b.ID = id
	s.data[id] = b
	return b, nil
}

func (s *UserDB) Patch(id int, partial map[string]any) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.data[id]
	if !ok {
		return User{}, errors.New("not found")
	}
	// very small/controlled patch
	if v, ok := partial["username"].(string); ok && v != "" {
		existing.Username = v
	}
	if v, ok := partial["email"].(string); ok && v != "" {
		existing.Email = v
	}
	if v, ok := partial["phone"].(string); ok && v != "" {
		existing.Phone = v
	}
	s.data[id] = existing
	return existing, nil
}

func (s *UserDB) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return errors.New("not found")
	}
	delete(s.data, id)
	return nil
}

type Server struct {
	store *UserDB
}

func NewServer() *Server {
	return &Server{store: NewUserDB()}
}

func (sv *Server) listUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, sv.store.List())
}

func (sv *Server) getUser(c echo.Context) error {
	id, err := atoiParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	user, err := sv.store.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "book not found")
	}
	return c.JSON(http.StatusOK, user)
}

func (sv *Server) createUser(c echo.Context) error {
	var in User
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
	}
	if in.Username == "" || in.Email == "" || in.Phone == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title and author are required")
	}
	created := sv.store.Create(in)
	c.Response().Header().Set("Location", fmt.Sprintf("/users/%d", created.ID))
	return c.JSON(http.StatusCreated, created)
}

func (sv *Server) updateUser(c echo.Context) error {
	//fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\n")
	//fmt.Println("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB\n")
	id, err := atoiParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var in User
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
	}
	fmt.Printf("Parsed struct: %+v\n", in)
	data, _ := json.MarshalIndent(in, "", "  ")
	fmt.Println("Input JSON:\n", string(data))
	if in.Username == "" || in.Email == "" || in.Phone == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Username, Email, and Phone number are required")
	}
	updated, uerr := sv.store.Update(id, in)
	if uerr != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	return c.JSON(http.StatusOK, updated)
}

func (sv *Server) patchUser(c echo.Context) error {
	id, err := atoiParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var patch map[string]any
	if err := c.Bind(&patch); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
	}
	updated, uerr := sv.store.Patch(id, patch)
	if uerr != nil {
		return echo.NewHTTPError(http.StatusNotFound, "book not found")
	}
	return c.JSON(http.StatusOK, updated)
}

func (sv *Server) deleteUser(c echo.Context) error {
	id, err := atoiParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := sv.store.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "book not found")
	}
	return c.NoContent(http.StatusNoContent)
}

func atoiParam(c echo.Context, name string) (int, error) {
	return strconv.Atoi(c.Param(name))
}

func main() {
	e := echo.New()

	// Middlewares (pretty standard defaults)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check
	e.GET("/healthz", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	// Routes
	sv := NewServer()
	g := e.Group("/users")
	g.GET("", sv.listUsers)         // GET /users
	g.GET("/:id", sv.getUser)       // GET /users/:id
	g.POST("", sv.createUser)       // POST /users
	g.PUT("/:id", sv.updateUser)    // PUT /users/:id
	g.PATCH("/:id", sv.patchUser)   // PATCH /users/:id
	g.DELETE("/:id", sv.deleteUser) // DELETE /users/:id

	// Graceful shutdown
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(e.AcquireContext().Request().Context(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
