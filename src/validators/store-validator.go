package validators

import (
	"product/src/dtos"
	"product/src/models"

	"github.com/google/uuid"

	common_validator "github.com/JohnSalazar/microservices-go-common/validators"
)

type addStore struct {
	ProductID uuid.UUID `from:"productid" json:"productid" validate:"required"`
	Quantity  uint      `from:"quantity" json:"quantity" validate:"required,gte=1"`
}

type bookStore struct {
	Products []*models.Product `from:"products" json:"products" validate:"required"`
}

type unbookStore struct {
	ID uuid.UUID `from:"id" json:"id" validate:"required"`
}

type paymentStore struct {
	ID   uuid.UUID `from:"id" json:"id" validate:"required"`
	Sold bool      `from:"sold" json:"sold" validate:"required"`
}

func ValidateAddStore(fields *dtos.AddStore) interface{} {
	addStore := addStore{
		ProductID: fields.ProductID,
		Quantity:  fields.Quantity,
	}

	err := common_validator.Validate(addStore)
	if err != nil {
		return err
	}

	return nil
}

func ValidateBookStore(fields *dtos.BookStore) interface{} {
	bookStore := bookStore{
		Products: fields.Products,
	}

	err := common_validator.Validate(bookStore)
	if err != nil {
		return err
	}

	return nil
}

func ValidateUnbookStore(fields *dtos.UnbookStore) interface{} {
	unbookStore := unbookStore{
		ID: fields.ID,
	}

	err := common_validator.Validate(unbookStore)
	if err != nil {
		return err
	}

	return nil
}

func ValidatePaymentStore(fields *dtos.PaymentStore) interface{} {
	paymentStore := paymentStore{
		ID:   fields.ID,
		Sold: fields.Sold,
	}

	err := common_validator.Validate(paymentStore)
	if err != nil {
		return err
	}

	return nil
}
