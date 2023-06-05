package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	command_product "product/src/application/commands/product"
	product_command_handler "product/src/application/commands/product/postgres"
	"product/src/models"
)

func Run(command_handler *product_command_handler.ProductCommandHandler) {

	err := handler(command_handler)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Seed done!!!")
}

func handler(command *product_command_handler.ProductCommandHandler) error {
	ctx := context.Background()

	jsonFile, err := os.Open(`./data/product.json`)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValueJSON, _ := io.ReadAll(jsonFile)

	products := []*models.Product{}
	err = json.Unmarshal(byteValueJSON, &products)
	if err != nil {
		return err
	}

	for _, product := range products {
		createProductCommand := &command_product.CreateProductCommand{
			ID:          product.ID,
			Name:        product.Name,
			Slug:        product.Slug,
			Description: product.Description,
			Price:       product.Price,
			Quantity:    product.Quantity,
			Image:       product.Image,
		}
		_, err = command.CreateProductCommandHandler(ctx, createProductCommand)
		if err != nil {
			return err
		}
	}

	return nil
}
