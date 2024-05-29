package merchant_entity

type MerchantCategories string

const (
	SmallRestaurant       MerchantCategories = "SmallRestaurant"
	MediumRestaurant      MerchantCategories = "MediumRestaurant"
	LargeRestaurant       MerchantCategories = "LargeRestaurant"
	MerchandiseRestaurant MerchantCategories = "MerchandiseRestaurant"
	BoothKiosk            MerchantCategories = "BoothKiosk"
	ConvenienceStore      MerchantCategories = "ConvenienceStore"
)

type Location struct {
	Lat  float64 `json:"lat" validate:"required"`
	Long float64 `json:"long" validate:"required"`
}

type Merchant struct {
	Id       string             `json:"id"`
	Name     string             `json:"name"`
	Category MerchantCategories `json:"merchantCategory"`
	ImageURL string             `json:"imageUrl"`
	Location Location           `json:"location"`
	UserId   string             `json:"userId"`
}

type AddMerchantRequest struct {
	Name     string             `json:"name" validate:"required,min=2,max=30"`
	Category MerchantCategories `json:"merchantCategory" validate:"oneof='SmallRestaurant' 'MediumRestaurant' 'LargeRestaurant' 'MerchandiseRestaurant' 'BoothKiosk' 'ConvenienceStore'"`
	ImageURL string             `json:"imageUrl" validate:"required,imageurl"`
	Location Location           `json:"location"`
}

type AddMerchantResponse struct {
	Id string `json:"merchantId"`
}
