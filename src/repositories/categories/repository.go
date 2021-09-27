package categories

import (
	"context"
	"fmt"
	"time"
	"ugc_test_task/src/models"
	pg2 "ugc_test_task/src/pg"

	sql "github.com/huandu/go-sqlbuilder"
)

const (
	TableName = "categories"
)

var (
	categoryFields = []string{models.IdKey, models.NameKey, models.CreateAt}
	indexFields    = []string{models.NameKey, models.CreateAt}
)

type Repository struct {
	client pg2.Client
}

func New(conf Config) (r Repository, err error) {
	if err := conf.Validate(); err != nil {
		return r, fmt.Errorf("config is invalid: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r.client, err = pg2.Connect(ctx, conf.pgConfig)
	if err != nil {
		return Repository{}, err
	}
	if err := r.createTable(); err != nil {
		return Repository{}, fmt.Errorf("create '%s' table: %v", TableName, err)
	}
	if err := r.createIndexes(); err != nil {
		return Repository{}, err
	}
	return r, nil
}

func (r Repository) createTable() error {
	s := sql.CreateTable(TableName).IfNotExists().
		Define(models.IdKey, "uuid", "primary key", "not null").
		Define(models.NameKey, "ltree", fmt.Sprintf("check (%s != '')", models.NameKey)).
		Define(models.CreateAt, "bigint", fmt.Sprintf("check (%s > 0)", models.CreateAt)).String()
	_, err := r.client.Exec(context.Background(), s)
	if err != nil {
		return err
	}
	return nil
}

func (r Repository) createIndexes() error {
	for _, indexField := range indexFields {
		indexType := "btree"
		if indexField == models.NameKey {
			indexType = "gist"
		}
		sqlStr := fmt.Sprintf("create index if not exists %s_idx on %s using %s (%s)", indexField, TableName, indexType, indexField)
		_, err := r.client.Exec(context.Background(), sqlStr)
		if err != nil {
			return fmt.Errorf("create index for field '%s': %v", indexField, err)
		}
	}
	return nil
}

func (r Repository) Insert(ctx context.Context, category models.Category) error {
	//todo: handle error
	if err := category.Validate(); err != nil {
		return err
	}
	sqlStr, args := sql.InsertInto(TableName).Cols(categoryFields...).
		Values(category.Id, category.Name, category.CreateAt).BuildWithFlavor(sql.PostgreSQL)
	if _, err := r.client.Exec(ctx, sqlStr, args...); err != nil {
		return pg2.NewError(err)
	}
	return nil
}

func (r Repository) IsEmpty() bool {
	return r.client.IsEmpty()
}