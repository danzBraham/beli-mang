package http

import (
	"log"
	"net/http"

	http_helper "github.com/danzBraham/beli-mang/internal/helpers/http"
	validator_helper "github.com/danzBraham/beli-mang/internal/helpers/validator"
	"github.com/danzBraham/beli-mang/internal/http/controllers"
	"github.com/danzBraham/beli-mang/internal/http/middlewares"
	"github.com/danzBraham/beli-mang/internal/repositories"
	"github.com/danzBraham/beli-mang/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type APIServer struct {
	Addr string
	DB   *pgxpool.Pool
}

func NewAPIServer(addr string, db *pgxpool.Pool) *APIServer {
	return &APIServer{
		Addr: addr,
		DB:   db,
	}
}

func (s *APIServer) Launch() error {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Beli Mang API"))
	})

	validator_helper.InitCustomValidation()

	// User domain
	userRepository := repositories.NewUserRepository(s.DB)
	userService := services.NewUserService(userRepository)
	userController := controllers.NewUserController(userService)
	adminController := controllers.NewAdminController(userService)

	// Merchant domain
	merchantRepository := repositories.NewMerchantRepository(s.DB)
	merchantService := services.NewMerchantService(merchantRepository)
	merchantController := controllers.NewMerchantController(merchantService)

	// Item domain
	itemRepository := repositories.NewItemRepository(s.DB)
	itemService := services.NewItemService(itemRepository, merchantRepository)
	itemController := controllers.NewItemController(itemService)

	// Purchase domain
	purchaseRepository := repositories.NewPurchaseRepository(s.DB)
	purchaseService := services.NewPurchaseService(purchaseRepository, merchantRepository, itemRepository)
	purchaseController := controllers.NewPurchaseController(purchaseService)

	r.Route("/admin", func(r chi.Router) {
		r.Mount("/", adminController.Routes())
		r.Mount("/merchants", merchantController.Routes())
		r.Mount("/merchants/{merchantId}/items", itemController.Routes())
	})

	r.Mount("/merchants", purchaseController.MerchantRoutes())

	r.Route("/users", func(r chi.Router) {
		r.Mount("/", userController.Routes())
		r.Group(func(r chi.Router) {
			r.Use(middlewares.Authenticate)
			r.Post("/estimate", purchaseController.HandleUserEstimateOrder)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http_helper.ResponseError(w, http.StatusNotFound, "Not found error", "Route does not exists")
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http_helper.ResponseError(w, http.StatusMethodNotAllowed, "Method not allowed error", "Method is not allowed")
	})

	server := http.Server{
		Addr:    s.Addr,
		Handler: r,
	}

	log.Printf("Server listening on %s\n", s.Addr)
	return server.ListenAndServe()
}
