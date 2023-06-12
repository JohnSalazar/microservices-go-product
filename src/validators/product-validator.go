package validators

import (
	"product/src/dtos"

	common_validator "github.com/JohnSalazar/microservices-go-common/validators"
	"github.com/google/uuid"
)

type addProduct struct {
	Name        string  `from:"name" json:"name" validate:"required,max=500"`
	Slug        string  `from:"slug" json:"slug" validate:"required,max=600"`
	Description string  `from:"description" json:"description,omitempty" validate:"max=10000"`
	Price       float32 `from:"price" json:"price" validate:"required,gte=1"`
	Quantity    uint    `from:"quantity" json:"quantity" validate:"required,gte=1"`
}

type updateProduct struct {
	ID          uuid.UUID `from:"id" json:"id" validate:"required"`
	Name        string    `from:"name" json:"name" validate:"max=500"`
	Slug        string    `from:"slug" json:"slug" validate:"required,max=600"`
	Description string    `from:"description" json:"description,omitempty" validate:"max=10000"`
	Price       float32   `from:"price" json:"price" validate:"required,gte=1"`
}

func ValidateAddProduct(fields *dtos.AddProduct) interface{} {
	addProduct := addProduct{
		Name: fields.Name,
		Slug: fields.Slug,
		// Description: fields.Description,
		Price:    fields.Price,
		Quantity: fields.Quantity,
	}

	err := common_validator.Validate(addProduct)
	if err != nil {
		return err
	}

	return nil
}

func ValidateUpdateProduct(fields *dtos.UpdateProduct) interface{} {
	updateProduct := updateProduct{
		ID:   fields.ID,
		Name: fields.Name,
		Slug: fields.Slug,
		// Description: fields.Description,
		Price: fields.Price,
	}

	err := common_validator.Validate(updateProduct)
	if err != nil {
		return err
	}

	return nil
}
