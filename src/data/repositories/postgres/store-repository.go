package postgres_repository

import (
	"context"
	"database/sql"
	"fmt"
	"product/src/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

type storeRepository struct {
	database *sql.DB
}

func NewStoreRepository(database *sql.DB) *storeRepository {
	return &storeRepository{
		database: database,
	}
}

func (r *storeRepository) LoadBookedStore(ctx context.Context) ([]*models.Store, error) {
	rows, err := r.database.QueryContext(ctx, `SELECT 
																							id,
																							productid, 
																							COALESCE(booked_at, '1900-01-01 00:00') booked_at,
																							sold,
																							created_at,
																							COALESCE(updated_at, '1900-01-01 00:00') updated_at,
																							version
																						FROM stores 
																						WHERE 
																							sold = false 
																							AND deleted = false 
																							AND booked_at >= $1`, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []*models.Store
	for rows.Next() {
		var store *models.Store
		err = rows.Scan(
			&store.ID,
			&store.ProductID,
			&store.BookedAt,
			&store.Sold,
			&store.CreatedAt,
			&store.UpdatedAt,
			&store.Version)
		if err != nil {
			return nil, err
		}

		stores = append(stores, store)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return stores, nil
}

func (r *storeRepository) FindByID(ctx context.Context, ID uuid.UUID) (*models.Store, error) {
	var store models.Store
	row := r.database.QueryRowContext(ctx, `SELECT
																						id,
																						productid, 
																						COALESCE(booked_at, '1900-01-01 00:00') booked_at,
																						sold,
																						created_at,
																						COALESCE(updated_at, '1900-01-01 00:00') updated_at,
																						version 
																					FROM stores 
																					WHERE 
																						deleted = false 
																						AND id = $1`, ID)
	if err := row.Scan(
		&store.ID,
		&store.ProductID,
		&store.BookedAt,
		&store.Sold,
		&store.CreatedAt,
		&store.UpdatedAt,
		&store.Version); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &store, nil
}

func (r *storeRepository) Book(ctx context.Context, productID uuid.UUID, quantity uint) ([]*models.Store, error) {
	rows, err := r.database.QueryContext(ctx, `SELECT 
																								id,
																								productid, 
																								COALESCE(booked_at, '1900-01-01 00:00') booked_at,
																								sold,
																								created_at,
																								COALESCE(updated_at, '1900-01-01 00:00') updated_at,
																								version 
																							FROM stores 
																							WHERE 
																								deleted = false 
																								AND sold = false 
																								AND productid = $1 
																								LIMIT $2`, productID.String(), quantity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []*models.Store
	for rows.Next() {
		var store models.Store
		err = rows.Scan(
			&store.ID,
			&store.ProductID,
			&store.BookedAt,
			&store.Sold,
			&store.CreatedAt,
			&store.UpdatedAt,
			&store.Version)
		if err != nil {
			return nil, err
		}

		stores = append(stores, &store)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return stores, nil
}

func (r *storeRepository) Create(ctx context.Context, stores []*models.Store) error {
	var (
		params []string
		vals   []interface{}
	)

	for i := 0; i < len(stores); i++ {
		params = append(params, fmt.Sprintf("($%v,$%v,$%v,$%v,$%v,$%v,$%v)",
			i*7+1,
			i*7+2,
			i*7+3,
			i*7+4,
			i*7+5,
			i*7+6,
			i*7+7,
		))
		vals = append(vals,
			stores[i].ID,
			stores[i].ProductID,
			stores[i].BookedAt,
			stores[i].Sold,
			stores[i].CreatedAt,
			stores[i].Version,
			stores[i].Deleted)
	}

	statement := fmt.Sprintf(`INSERT INTO stores (
													id, 
													productid, 
													booked_at,
													sold,
													created_at, 
													version,
													deleted) VALUES %s`, strings.Join(params, ","))

	_, err := r.database.ExecContext(ctx, statement, vals...)
	if err != nil {
		return err
	}

	return nil
}

func (r *storeRepository) Update(ctx context.Context, stores []*models.Store) ([]*models.Store, error) {
	var (
		params []string
		vals   []interface{}
	)

	for i := 0; i < len(stores); i++ {
		params = append(params, fmt.Sprintf("($%v,$%v,$%v,$%v,$%v)",
			i*5+1,
			i*5+2,
			i*5+3,
			i*5+4,
			i*5+5,
		))
		vals = append(vals,
			stores[i].ID,
			stores[i].BookedAt,
			stores[i].Sold,
			stores[i].UpdatedAt,
			stores[i].Version)
	}

	statement := fmt.Sprintf(`UPDATE stores SET 
															booked_at = s.booked_at::timestamp, 
															sold = s.sold::boolean, 
															updated_at = s.updated_at::timestamp,
															version = s.version::integer 
														FROM (VALUES %s) AS s(id,booked_at,sold,updated_at,version)
														WHERE 
															stores.id = s.id::uuid
															AND stores.version = s.version::integer-1`, strings.Join(params, ","))

	_, err := r.database.ExecContext(ctx, statement, vals...)
	if err != nil {
		return nil, err
	}

	return stores, nil
}

func (r *storeRepository) Delete(ctx context.Context, ID uuid.UUID) error {
	_, err := r.database.ExecContext(ctx, "UPDATE stores SET deleted = true WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
