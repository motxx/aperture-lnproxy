package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type DB struct {
	*bun.DB
}

type Content struct {
	bun.BaseModel  `bun:"table:contents"`
	Id             string `bun:"id,pk"`
	Title          string `bun:"title"`
	Author         string `bun:"author"`
	Filepath       string `bun:"filepath"`
	RecipientLud16 string `bun:"recipient_lud16"`
	Price          int64  `bun:"price"`
}

func NewDB(dataSourceName string) (*DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dataSourceName)))
	db := bun.NewDB(sqldb, pgdialect.New())
	return &DB{db}, nil
}

func (d *DB) AddContent(content *Content) (string, error) {
	_, err := d.NewInsert().Model(content).Exec(context.Background())
	if err != nil {
		return "", err
	}
	return content.Id, nil
}

func (d *DB) UpdateContent(content *Content) (string, error) {
	_, err := d.NewUpdate().Model(content).Where("id = ?", content.Id).Exec(context.Background())
	if err != nil {
		return "", err
	}
	return content.Id, nil
}

func (d *DB) RemoveContent(id string) (string, error) {
	count, err := d.NewSelect().Model((*Content)(nil)).Where("id = ?", id).Count(context.Background())
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "", fmt.Errorf("content with id %s does not exist", id)
	}

	_, err = d.NewDelete().Model((*Content)(nil)).Where("id = ?", id).Exec(context.Background())
	if err != nil {
		return "", err
	}
	return id, nil
}

func (d *DB) GetContent(id string) (*Content, error) {
	content := new(Content)
	err := d.NewSelect().Model(content).Where("id = ?", id).Scan(context.Background())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no content with id %s", id)
		}
		return nil, err
	}
	return content, nil
}
