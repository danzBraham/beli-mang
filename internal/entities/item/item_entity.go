package item_entity

const (
	Beverage   string = "Beverage"
	Food       string = "Food"
	Snack      string = "Snack"
	Condiments string = "Condiments"
	Additions  string = "Additions"
)

type Item struct {
	Id         string
	Name       string
	Category   string
	Price      int
	ImageURL   string
	MerchantId string
	CreatedAt  string
	UpdatedAt  string
}

type AddItemRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=30"`
	Category string `json:"productCategory" validate:"oneof='Beverage' 'Food' 'Snack' 'Condiments' 'Additions'"`
	Price    int    `json:"price" validate:"required,min=1"`
	ImageURL string `json:"imageUrl" validate:"required,imageurl"`
}

type AddItemResponse struct {
	Id string `json:"itemId"`
}

type ItemQueryParams struct {
	Id        string
	Limit     int
	Offset    int
	Name      string
	Category  string
	CreatedAt string
}

type GetItem struct {
	Id        string `json:"itemId"`
	Name      string `json:"name"`
	Category  string `json:"productCategory"`
	Price     int    `json:"price"`
	ImageURL  string `json:"imageUrl"`
	CreatedAt string `json:"createdAt"`
}

type Meta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type GetItemResponse struct {
	Data []*GetItem `json:"data"`
	Meta Meta       `json:"meta"`
}
