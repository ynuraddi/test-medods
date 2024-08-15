package postgres

import (
	"context"
	"database/sql"
	"medods/internal/model"
)

type User struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) *User {
	return &User{conn: conn}
}

func (r User) Create(ctx context.Context, u model.User) error {
	query := `insert into users(email) values($1)`
	_, err := r.conn.ExecContext(ctx, query, u.Email)
	return err
}

func (r User) GetByID(ctx context.Context, id int) (u model.User, err error) {
	query := `select id, email from users where id = $1`
	err = r.conn.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
	)
	return u, err
}

func (r User) List(ctx context.Context) ([]model.User, error) {
	query := `select id, email from users`
	rows, err := r.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID,
			&u.Email,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
