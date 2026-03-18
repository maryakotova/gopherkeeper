package main

import (
	"GophKeeper/server/internal/auth"
	"GophKeeper/server/internal/config"
	"GophKeeper/server/internal/handlers"
	"GophKeeper/server/internal/service"
	"GophKeeper/server/internal/storage"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	err := run()
	if err != nil {
		log.Printf("Ошибка при запуске сервера: %v", err)
		os.Exit(1)
	}
}

func run() error {

	config := config.NewConfig()

	factory := &storage.StorageFactory{}

	repo, err := factory.NewStorage(config.DatabaseURI)
	if err != nil {
		return err
	}

	defer func() {
		err = repo.Close()
		if err != nil {
			log.Printf("Ошибка при закрытии соединения БД: %v", err)
		}
	}()

	authService := service.NewAuthService(repo)
	controllerService := service.NewControllerService(repo)

	authHandlers := handlers.NewAuthHandler(authService)
	controllerHandlers := handlers.NewControllerHandler(controllerService)

	r := setupRoutes(authHandlers, controllerHandlers)
	r.Run(config.RunAddress)

	return nil
}

func setupRoutes(authHandler *handlers.AuthHandler, controller *handlers.ControllerHandler) *gin.Engine {

	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.MaxMultipartMemory = 2 << 30

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	api := r.Group("/api")
	{
		api.POST("/register", authHandler.RegisterHandler)
		api.POST("/login", authHandler.LoginHandler)
	}

	autorized := r.Group("/data")
	{
		autorized.Use(auth.AuthMiddleware())
		autorized.GET("/download/:id", controller.DownloadHandler)
		autorized.POST("/upload", controller.UploadHandler)
		autorized.POST("/update", controller.UpdateHandler)
		autorized.GET("/sync/:lastsync", controller.SyncHandler)
	}

	return r
}
