package repository

import (
	"context"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	List(ctx context.Context) ([]User, error)
	Get(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, id int64, u *User) error
	Delete(ctx context.Context, id int64) error
}
