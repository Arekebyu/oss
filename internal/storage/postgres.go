package storage

import (
	"context"
	"fmt"
	"oss/internal/types"
	"time"

	// Import your crawler types

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(connString string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %v", err)
	}
	return &DB{Pool: pool}, nil
}

func (db *DB) SavePage(ctx context.Context, p types.ScrapedPage) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queryPage := `
		INSERT INTO pages (url, title, crawled_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (url)
		DO UPDATE SET title = EXCLUDED.title, crawled_at = EXCLUDED.crawled_AT
		RETURNING id;
		`
	var pageID int
	err = tx.QueryRow(ctx, queryPage, p.URL, p.Title, time.Now()).Scan(&pageID)
	if err != nil {
		return fmt.Errorf("failed to save page: %v", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM sections WHERE page_id = $1`, pageID)
	if err != nil {
		return fmt.Errorf("failed to delete old section: %v", err)
	}

	querySection := `
		INSERT INTO sections (page_id, section_type, content, language, sort_order)
		VALUES ($1, $2, $3, $4, $5)
	`

	for i, section := range p.Sections {
		_, err = tx.Exec(ctx, querySection,
			pageID,
			section.Type,
			section.Content,
			section.Language,
			i)
		if err != nil {
			fmt.Errorf("failed to save section %v with error %v", i, err)
		}
	}
	return tx.Commit(ctx)
}

func (db *DB) Close() {
	db.Pool.Close()
}
