package merchant_entity

const (
	SmallRestaurant       string = "SmallRestaurant"
	MediumRestaurant      string = "MediumRestaurant"
	LargeRestaurant       string = "LargeRestaurant"
	MerchandiseRestaurant string = "MerchandiseRestaurant"
	BoothKiosk            string = "BoothKiosk"
	ConvenienceStore      string = "ConvenienceStore"
)

type Location struct {
	Lat  float64 `json:"lat" validate:"required"`
	Long float64 `json:"long" validate:"required"`
}

type Merchant struct {
	Id        string
	Name      string
	Category  string
	ImageURL  string
	Location  Location
	UserId    string
	CreatedAt string
	UpdatedAt string
}

type AddMerchantRequest struct {
	Name     string   `json:"name" validate:"required,min=2,max=30"`
	Category string   `json:"merchantCategory" validate:"oneof='SmallRestaurant' 'MediumRestaurant' 'LargeRestaurant' 'MerchandiseRestaurant' 'BoothKiosk' 'ConvenienceStore'"`
	ImageURL string   `json:"imageUrl" validate:"required,imageurl"`
	Location Location `json:"location"`
}

type AddMerchantResponse struct {
	Id string `json:"merchantId"`
}

type MerchantQueryParams struct {
	Id        string
	Limit     int
	Offset    int
	Name      string
	Category  string
	CreatedAt string
}

type GetMerchant struct {
	Id        string   `json:"merchantId"`
	Name      string   `json:"name"`
	Category  string   `json:"merchantCategory"`
	ImageURL  string   `json:"imageUrl"`
	Location  Location `json:"location"`
	CreatedAt string   `json:"createdAt"`
}

type Meta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type GetMerchantResponse struct {
	Data []*GetMerchant `json:"data"`
	Meta Meta           `json:"meta"`
}
