package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/task"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/auth"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/balance"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/order"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/withdrawal"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	internalMiddleware "github.com/m1khal3v/gophermart-loyalty-service/internal/middleware"
	pkgMiddleware "github.com/m1khal3v/gophermart-loyalty-service/pkg/middleware"
	"time"
)

func New(
	appEnv string,
	userManager *manager.UserManager,
	orderManager *manager.OrderManager,
	taskManager *task.Manager,
	withdrawalManager *manager.WithdrawalManager,
	userWithdrawalManager *manager.UserWithdrawalManager,
	jwt *jwt.Container,
) chi.Router {
	authRoutes := auth.NewContainer(userManager)
	orderRoutes := order.NewContainer(orderManager, taskManager)
	balanceRoutes := balance.NewContainer(userManager, withdrawalManager, userWithdrawalManager)
	withdrawalRoutes := withdrawal.NewContainer(withdrawalManager)

	router := chi.NewRouter()
	router.Use(pkgMiddleware.ZapLogRequest(logger.Logger, "http-request"))
	router.Use(internalMiddleware.Recover())
	router.Use(pkgMiddleware.ZapLogPanic(logger.Logger, "http-panic"))
	router.Use(middleware.RealIP)
	router.Use(pkgMiddleware.Decompress())
	router.Use(pkgMiddleware.Compress(5, "text/html", "application/json"))
	router.Route("/api", func(router chi.Router) {
		router.Route("/user", func(router chi.Router) {
			// Anonymous
			router.Group(func(router chi.Router) {
				if appEnv == "prod" {
					router.Use(httprate.Limit(
						1,
						time.Second*3,
						httprate.WithKeyByRealIP(),
						httprate.WithLimitHandler(controller.RateLimited),
					))
				}

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
					router.Post("/withdraw", balanceRoutes.Withdraw)
				})
				router.Get("/withdrawals", withdrawalRoutes.List)
			})
		})
	})

	return router
}
