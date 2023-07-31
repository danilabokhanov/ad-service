package postgrespgx

import (
	"adservice/internal/user"
	"context"
	pgxLogrus "github.com/jackc/pgx-logrus"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
	log "github.com/sirupsen/logrus"
)

type PGXUsers struct {
	conn *pgx.Conn
}

const SelectUserById = "SELECT id, nickname, email FROM users WHERE id = $1;"

func (d *PGXUsers) SelectById(ctx context.Context, userID int64) ([]user.User, error) {
	rows, err := d.conn.Query(ctx, SelectUserById, userID)
	if err != nil {
		return []user.User{}, err
	}
	defer rows.Close()
	var res []user.User
	for rows.Next() {
		u := user.User{}
		err = rows.Scan(&u.ID, &u.Nickname, &u.Email)
		if err != nil {
			return []user.User{}, err
		}
		res = append(res, u)
	}
	return res, nil
}

const UpdateUserById = "UPDATE users SET nickname = $1, email = $2 WHERE id = $3;"

func (d *PGXUsers) UpdateById(ctx context.Context, userID int64, nickname, email string) error {
	rows, err := d.conn.Query(ctx, UpdateUserById, nickname, email, userID)
	defer rows.Close()
	return err
}

const InsertUser = "INSERT INTO users (nickname, email, id) VALUES ($1, $2, $3);"

func (d *PGXUsers) Insert(ctx context.Context, nickname string, email string, userID int64) error {
	rows, err := d.conn.Query(ctx, InsertUser, nickname, email, userID)
	defer rows.Close()
	return err
}

const DeleteUserById = "DELETE FROM users WHERE id = $1;"

func (d *PGXUsers) DeleteById(ctx context.Context, userID int64) error {
	rows, err := d.conn.Query(ctx, DeleteUserById, userID)
	defer rows.Close()
	return err
}

func (d *PGXUsers) Close() error {
	return d.conn.Close(context.Background())
}

const SelectAllUsers = "SELECT id, nickname, email FROM users;"

func (d *PGXUsers) SelectAll(ctx context.Context) ([]user.User, error) {
	rows, err := d.conn.Query(ctx,
		SelectAllUsers)
	if err != nil {
		return []user.User{}, err
	}
	defer rows.Close()
	res := []user.User{}
	for rows.Next() {
		u := user.User{}
		err = rows.Scan(&u.ID, &u.Nickname, &u.Email)
		if err != nil {
			return []user.User{}, err
		}
		res = append(res, u)
	}
	return res, nil
}

const DeleteAllUsers = "DELETE FROM users"

func (d *PGXUsers) DeleteAll(ctx context.Context) error {
	rows, err := d.conn.Query(ctx, DeleteAllUsers)
	defer rows.Close()
	return err
}

func NewDBUsers() PGXUsers {
	logger := log.New()
	logger.SetLevel(log.InfoLevel)
	logger.SetFormatter(&log.TextFormatter{})

	config, err := pgx.ParseConfig("postgres://postgres:postgres@localhost:5433/ad_service")
	if err != nil {
		logger.WithError(err).Fatalf("can't parse pgx config")
	}

	config.Tracer = &tracelog.TraceLog{
		Logger:   pgxLogrus.NewLogger(logger),
		LogLevel: tracelog.LogLevelDebug,
	}

	conn, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		logger.WithError(err).Fatalf("can't connect to pg")
	}
	return PGXUsers{conn: conn}
}
