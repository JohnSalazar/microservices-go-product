package postgres_repository

import (
	"context"
	"database/sql"
	"product/src/models"
	"time"

	"github.com/google/uuid"
)

type productRepository struct {
	database *sql.DB
}

func NewProductRepository(database *sql.DB) *productRepository {
	return &productRepository{
		database: database,
	}
}

func (r *productRepository) GetAll(ctx context.Context, name string, page int, size int) ([]*models.Product, error) {
	rows, err := r.database.QueryContext(ctx,
		`SELECT 
																					id, 
																					name,
																					slug, 
																					description, 
																					price,
																					image,
																					created_at, 
																					COALESCE(updated_at, '1900-01-01 00:00') updated_at, 
																					version 
																				FROM products 
																				WHERE name ILIKE '%' || $1 || '%'
																				AND deleted = false
																				ORDER BY name ASC
																				LIMIT $2 OFFSET $3`, name, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err = rows.Scan(
			&product.ID,
			&product.Name,
			&product.Slug,
			&product.Description,
			&product.Price,
			&product.Image,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.Version)
		if err != nil {
			return nil, err
		}

		products = append(products, &product)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepository) FindByID(ctx context.Context, ID uuid.UUID) (*models.Product, error) {
	var product models.Product
	row := r.database.QueryRowContext(
		ctx,
		`SELECT 
		id, 
		name, 
		slug, 
		description, 
		price,
		image,
		created_at, 
		COALESCE(updated_at, '1900-01-01 00:00') updated_at, 
		version
		FROM products WHERE id = $1`,
		ID,
	)
	if err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Image,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Version); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil

	// err := row.Scan(
	// 	&product.ID,
	// 	&product.Name,
	// 	&product.Slug,
	// 	&product.Description,
	// 	&product.Price,
	// 	&product.CreatedAt,
	// 	&product.UpdatedAt,
	// 	&product.Version)
	// switch err {
	// case sql.ErrNoRows:
	// 	return nil, nil
	// case nil:
	// 	return &product, nil
	// default:
	// 	return nil, err
	// }
}

func (r *productRepository) FindBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	row := r.database.QueryRowContext(
		ctx,
		`SELECT 
		id, 
		name, 
		slug, 
		description, 
		price,
		image,
		created_at, 
		COALESCE(updated_at, '1900-01-01 00:00') updated_at, 
		version,
		(	SELECT COUNT(productid) 
			FROM stores 
			WHERE productid = products.id 
			AND stores.deleted = false 
			AND sold = false
			AND booked_at <= NOW()::timestamptz
			) as quantity
		FROM products WHERE slug = $1`,
		slug,
	)
	if err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Image,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Version,
		&product.Quantity); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) FindByName(ctx context.Context, name string) (*models.Product, error) {
	var product models.Product
	row := r.database.QueryRowContext(ctx, "SELECT id, name, slug, description, price, image, created_at, COALESCE(updated_at, '1900-01-01 00:00') updated_at, version FROM products WHERE name = $1", name)
	if err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Image,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Version); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	sql := "INSERT INTO products (id, name, slug, description, price, image, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"

	_, err := r.database.ExecContext(ctx, sql,
		product.ID,
		product.Name,
		product.Slug,
		product.Description,
		product.Price,
		product.Image,
		product.CreatedAt)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) (*models.Product, error) {
	sql := "UPDATE products SET name = $1, slug = $2, description = $3, price = $4, image = $5, updated_at = $6, version = $7 WHERE id = $8 and version = ($7-1)"

	product.Version++
	product.UpdatedAt = time.Now().UTC()
	_, err := r.database.ExecContext(ctx, sql,
		product.Name,
		product.Slug,
		product.Description,
		product.Price,
		product.Image,
		product.UpdatedAt,
		product.Version,
		product.ID)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) Delete(ctx context.Context, ID uuid.UUID) error {
	_, err := r.database.ExecContext(ctx, "UPDATE products SET deleted = true WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
