package user_entity

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"isAdmin"`
}

type RegisterAdminUserRequest struct {
	Username string `json:"username" validate:"required,min=5,max=30"`
	Password string `json:"password" validate:"required,min=5,max=30"`
	Email    string `json:"email" validate:"required,email"`
}

type RegisterAdminUserResponse struct {
	Token string `json:"token"`
}
