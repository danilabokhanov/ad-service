package postgrespgx

import (
	"adservice/internal/ads"
	"context"
	pgxLogrus "github.com/jackc/pgx-logrus"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
	log "github.com/sirupsen/logrus"
	"time"
)

type PGXRepo struct {
	conn *pgx.Conn
}

const SelectAdById = "SELECT id, title, text, author_id, published, creation_date, update_date " +
	"FROM ads WHERE id = $1;"

func (d *PGXRepo) SelectById(ctx context.Context, adID int64) ([]ads.Ad, error) {
	rows, err := d.conn.Query(ctx, SelectAdById, adID)
	defer rows.Close()
	var res []ads.Ad
	for rows.Next() {
		ad := ads.Ad{}
		var creationDate, updateDate int64
		err = rows.Scan(&ad.ID, &ad.Title, &ad.Text, &ad.AuthorID, &ad.Published, &creationDate, &updateDate)
		ad.CreationDate = time.UnixMicro(creationDate).UTC()
		ad.UpdateDate = time.UnixMicro(updateDate).UTC()
		if err != nil {
			return []ads.Ad{}, err
		}
		res = append(res, ad)
	}
	return res, nil
}

const SelectAllAdsIndexes = "SELECT id FROM ads;"

func (d *PGXRepo) SelectAllIndexes(ctx context.Context) ([]int64, error) {
	rows, err := d.conn.Query(ctx, SelectAllAdsIndexes)
	if err != nil {
		return []int64{}, err
	}
	defer rows.Close()
	var res []int64
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return []int64{}, err
		}
		res = append(res, id)
	}
	return res, nil
}

const SelectAllAds = "SELECT id, title, text, author_id, published, creation_date, update_date FROM ads;"

func (d *PGXRepo) SelectAll(ctx context.Context) ([]ads.Ad, error) {
	rows, err := d.conn.Query(ctx, SelectAllAds)
	if err != nil {
		return []ads.Ad{}, err
	}
	defer rows.Close()
	var res []ads.Ad
	for rows.Next() {
		ad := ads.Ad{}
		var creationDate, updateDate int64
		err = rows.Scan(&ad.ID, &ad.Title, &ad.Text, &ad.AuthorID, &ad.Published, &creationDate, &updateDate)
		ad.CreationDate = time.UnixMicro(creationDate).UTC()
		ad.UpdateDate = time.UnixMicro(updateDate).UTC()
		if err != nil {
			return []ads.Ad{}, err
		}
		res = append(res, ad)
	}
	return res, nil
}

const InsertAd = "INSERT INTO ads (id, title, text, author_id, published, creation_date, update_date) VALUES" +
	"($1, $2, $3, $4, $5, $6, $7);"

func (d *PGXRepo) Insert(ctx context.Context, ad ads.Ad) error {
	rows, err := d.conn.Query(ctx, InsertAd, ad.ID, ad.Title, ad.Text, ad.AuthorID,
		ad.Published, ad.CreationDate.UnixMicro(), ad.UpdateDate.UnixMicro())
	defer rows.Close()
	return err
}

const UpdateAdTitle = "UPDATE ads SET title = $1, update_date = $2 WHERE id = $3;"

func (d *PGXRepo) UpdateTitle(ctx context.Context, adID int64, title string, UpdateDate time.Time) error {
	rows, err := d.conn.Query(ctx, UpdateAdTitle, title, UpdateDate.UnixMicro(), adID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var insertedRows uint
		err = rows.Scan(insertedRows)
		if err != nil {
			return err
		}
	}
	return err
}

const UpdateAdText = "UPDATE ads SET text = $1, update_date = $2 WHERE id = $3;"

func (d *PGXRepo) UpdateText(ctx context.Context, adID int64, text string, UpdateDate time.Time) error {
	rows, err := d.conn.Query(ctx, UpdateAdText, text, UpdateDate.UnixMicro(), adID)
	defer rows.Close()
	return err
}

const UpdateAdStatus = "UPDATE ads SET published = $1, update_date = $2 WHERE id = $3;"

func (d *PGXRepo) UpdateStatus(ctx context.Context, adID int64, status bool, UpdateDate time.Time) error {
	rows, err := d.conn.Query(ctx, UpdateAdStatus, status, UpdateDate.UnixMicro(), adID)
	defer rows.Close()
	return err
}

const DeleteAdById = "DELETE FROM ads WHERE id = $1;"

func (d *PGXRepo) DeleteById(ctx context.Context, adID int64) error {
	rows, err := d.conn.Query(ctx, DeleteAdById, adID)
	defer rows.Close()
	return err
}

const DeleteAdByAuthor = "DELETE FROM ads WHERE author_id = $1;"

func (d *PGXRepo) DeleteByAuthor(ctx context.Context, userID int64) error {
	rows, err := d.conn.Query(ctx, DeleteAdByAuthor, userID)
	defer rows.Close()
	return err
}

const DeleteAllAds = "DELETE FROM ads;"

func (d *PGXRepo) DeleteAll(ctx context.Context) error {
	rows, err := d.conn.Query(ctx, DeleteAllAds)
	defer rows.Close()
	return err
}

func NewDBRepo() PGXRepo {
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
	return PGXRepo{conn: conn}
}

func (d *PGXRepo) Close() error {
	return d.conn.Close(context.Background())
}
