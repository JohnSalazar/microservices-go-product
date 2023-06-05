package subjects

type ProductSubject string
type StoreSubject string

const (
	ProductCreateMongo    ProductSubject = "product:create-mongo"
	ProductCreatePostgres ProductSubject = "product:create-postgres"
	ProductUpdateMongo    ProductSubject = "product:update-mongo"
	StoreBookMongo        StoreSubject   = "store:book-mongo"
	StoreCreateMongo      StoreSubject   = "store:create-mongo"
	StoreCreatePostgres   StoreSubject   = "store:create-postgres"
	StorePaymentMongo     StoreSubject   = "store:payment-mongo"
	StorePaymentPostgres  StoreSubject   = "store:payment-postgres"
	StoreUnbookMongo      StoreSubject   = "store:unbook-mongo"
	StoreUnbookPostgres   StoreSubject   = "store:unbook-postgres"
)

func GetProductSubjects() []string {
	return []string{
		string(ProductCreateMongo),
		string(ProductCreatePostgres),
		string(ProductUpdateMongo),
	}
}

func GetStoreSubjects() []string {
	return []string{
		string(StoreBookMongo),
		string(StoreCreateMongo),
		string(StoreCreatePostgres),
		string(StorePaymentMongo),
		string(StorePaymentPostgres),
		string(StoreUnbookMongo),
		string(StoreUnbookPostgres),
	}
}
