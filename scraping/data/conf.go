package data

import (
	"context"
	"database/sql"
	"usdt/scraping/sqlc-config/config"
)

func NewConfig(db *sql.DB) (*config.Config, error) {
	ctx := context.Background()
	queries := config.New(db)
	c, err := queries.GetConfig(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return &c, nil
		}
		return &c, err
	}
	return &c, nil
}
