package postgres

import (
	"context"
	"database/sql"
	"medods/internal/model"
)

type Session struct {
	conn *sql.DB
}

func NewSessionRepository(conn *sql.DB) *Session {
	return &Session{
		conn: conn,
	}
}

func (r Session) Create(ctx context.Context, session model.Session) error {
	query := `
	insert into sessions(
		user_id,
		access_token_id,
		refresh_token_hash,
		created_at
	) values($1, $2, $3, $4)`

	_, err := r.conn.ExecContext(ctx, query,
		session.UserID,
		session.ATokenID,
		session.RTokenHash,
		session.CreatedAt,
	)
	return err
}

func (r Session) Update(ctx context.Context, session model.Session) (model.Session, error) {
	query := `
	update sessions set
		user_id = $3,
		access_token_id = $4,
		refresh_token_hash = $5,
		created_at = $6,
		version = version + 1
	where id = $1 and version = $2
	returning
		id, 
		user_id, 
		access_token_id, 
		refresh_token_hash, 
		created_at,
		version`

	var res model.Session
	err := r.conn.QueryRowContext(ctx, query,
		session.ID,
		session.Version,
		session.UserID,
		session.ATokenID,
		session.RTokenHash,
		session.CreatedAt,
	).Scan(
		&res.ID,
		&res.UserID,
		&res.ATokenID,
		&res.RTokenHash,
		&res.CreatedAt,
		&res.Version,
	)
	if err != nil {
		return model.Session{}, err
	}
	return res, nil
}

func (r Session) GetByUserID(ctx context.Context, id int) (model.Session, error) {
	query := `
	select
		id,
		user_id,
		access_token_id,
		refresh_token_hash,
		created_at,
		version
	from sessions
	where user_id = $1`

	var res model.Session
	err := r.conn.QueryRowContext(ctx, query, id).Scan(
		&res.ID,
		&res.UserID,
		&res.ATokenID,
		&res.RTokenHash,
		&res.CreatedAt,
		&res.Version,
	)
	if err != nil {
		return model.Session{}, err
	}
	return res, nil

}

func (r Session) List(ctx context.Context) ([]model.Session, error) {
	query := `
	select
		id,
		user_id,
		access_token_id,
		refresh_token_hash,
		created_at,
		version
	from sessions`

	rows, err := r.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.ATokenID,
			&s.RTokenHash,
			&s.CreatedAt,
			&s.Version,
		); err != nil {
			return nil, err
		}

		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}
