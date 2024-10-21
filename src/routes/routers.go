package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/myrachanto/entaingo/src/api/controller"
	"github.com/myrachanto/entaingo/src/api/repository"
	"github.com/myrachanto/entaingo/src/api/service"
)

// var passer echo.MiddlewareFunc

func ApiServer() {
	// Test database connection
	fmt.Println("routes ...................................................")
	err := repository.IndexRepo.Dbsetup()
	if err != nil {
		log.Fatal(err)
	}
	userRepo := repository.NewUserRepo()
	u := controller.NewUserController(service.NewUserService(userRepo))
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	router.GET("/healthy", HealthCheck)
	router.POST("/transaction", u.Create)
	router.GET("/transaction/:id", u.GetTransactions)

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file in routes ", err)
	}

	PORT := os.Getenv("PORT")
	srv := &http.Server{
		Addr:    PORT,
		Handler: router,
	}

	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Launch CancelOddTransactions in a goroutine
	go userRepo.CancelOddTransactions(ctx, wg)

	// Start the HTTP server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Handle system signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Cancel the context to notify goroutines
	cancel()

	// Shut down the HTTP server with a context
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("Server exited gracefully.")
}
