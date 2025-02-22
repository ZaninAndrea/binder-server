package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/ZaninAndrea/binder-server/internal/log"
	"github.com/ZaninAndrea/binder-server/internal/mongo"
	"github.com/ZaninAndrea/binder-server/internal/rest"
	"github.com/ZaninAndrea/binder-server/storage"
	"github.com/dyson/certman"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var mainLogger = log.Default().Service("main")

func main() {
	// Load variables from local .env file
	err := godotenv.Load()
	if err != nil {
		mainLogger.Warn("Did not load the .env file: ", err)
	}

	mongoUri := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DATABASE")
	db := mongo.Connect(mongoUri, mongoDatabase)
	defer db.Disconnect()

	// Setup the Blob Storage
	storageAccount := os.Getenv("BLOB_STORAGE_ACCOUNT")
	storageKey := os.Getenv("BLOB_STORAGE_KEY")
	imagesStorage, err := storage.NewBlobStorage(storageAccount, storageKey, "images")
	if err != nil {
		mainLogger.Error(err)
		panic(err)
	}

	// Setup the HTTP server allowing all CORS
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	rest.SetupRoutes(router, db, imagesStorage)

	// Serving the .well-known route to allow automatic
	// Let's Encrypt certificate renewal
	router.Static("/.well-known", "./.well-known")

	// Setup certman to automatically reload the SSL certificate
	if os.Getenv("SSL_CERT") != "" {
		cm, err := certman.New(os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"))
		if err != nil {
			mainLogger.Error(err)
			panic(err)
		}
		cm.Logger(mainLogger)
		if err := cm.Watch(); err != nil {
			mainLogger.Error(err)
		}

		// Listen on the port specified in the environment
		s := &http.Server{
			Addr:    ":" + os.Getenv("SSL_PORT"),
			Handler: router,
			TLSConfig: &tls.Config{
				GetCertificate: cm.GetCertificate,
			},
		}

		go (func() {
			mainLogger.Info("Listening on %s", s.Addr)
			if err := s.ListenAndServeTLS("", ""); err != nil {
				mainLogger.Error(err)
			}
		})()
	}

	err = router.Run()
	if err != nil {
		panic(err)
	}
}
