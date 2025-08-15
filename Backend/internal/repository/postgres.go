package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // registers driver as "pgx"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(dsn string) (*PostgresRepo, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// connection sanity check
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, username, name, email, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *PostgresRepo) Get(ctx context.Context, id int64) (User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, `SELECT id, username, name, email, created_at FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *PostgresRepo) Create(ctx context.Context, u *User) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO users(username, name, email) VALUES($1,$2,$3) RETURNING id, created_at`,
		u.Username, u.Name, u.Email,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *PostgresRepo) Update(ctx context.Context, id int64, u *User) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET username=$1, name=$2, email=$3 WHERE id=$4`, u.Username, u.Name, u.Email, id,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *PostgresRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *PostgresRepo) Close() error { return r.db.Close() }

// Optional: helper for DSN from discrete env vars
func DSN(host, user, pass, db string, port int) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, pass, host, port, db)
}
