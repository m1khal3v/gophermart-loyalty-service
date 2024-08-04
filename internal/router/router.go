package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/auth"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/balance"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/order"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/withdrawal"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	internalMiddleware "github.com/m1khal3v/gophermart-loyalty-service/internal/middleware"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"gorm.io/gorm"
	"time"
)

func New(db *gorm.DB, jwt *jwt.Container) chi.Router {
	userRepository := repository.NewUserRepository(db)
	withdrawalRepository := repository.NewWithdrawalRepository(db)
	authRoutes := auth.NewContainer(userRepository, jwt)
	orderRoutes := order.NewContainer(repository.NewOrderRepository(db))
	balanceRoutes := balance.NewContainer(userRepository, jwt, withdrawalRepository)
	withdrawalRoutes := withdrawal.NewContainer(withdrawalRepository)

	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	router.Route("/api", func(router chi.Router) {
		router.Route("/user", func(router chi.Router) {
			// Anonymous
			router.Group(func(router chi.Router) {
				router.Use(httprate.Limit(
					1,
					time.Second*3,
					httprate.WithKeyByRealIP(),
					httprate.WithLimitHandler(controller.RateLimited),
				))

				router.Post("/register", authRoutes.Register)
				router.Post("/login", authRoutes.Login)
			})

			// Authorized
			router.Group(func(router chi.Router) {
				router.Use(internalMiddleware.ValidateAuthorizationToken(jwt))

				router.Post("/orders", orderRoutes.Register)
				router.Get("/orders", orderRoutes.List)
				router.Route("/balance", func(router chi.Router) {
					router.Get("/", balanceRoutes.Balance)
				})
				router.Get("/withdrawals", withdrawalRoutes.List)
			})
		})
	})

	return router
}
