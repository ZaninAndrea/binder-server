package rest

import (
	"os"

	"github.com/ZaninAndrea/binder-server/internal/log"
	"github.com/ZaninAndrea/binder-server/internal/mongo"
	"github.com/gin-gonic/gin"
)

var restLogger = log.Default().Service("rest")

func SetupRoutes(r *gin.Engine, db *mongo.Database) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	r.Use(ParseAuthorizationHeader(jwtSecret))
	setupUserRoutes(r, db, jwtSecret)
	setupDeckRoutes(r, db)
	setupCardRoutes(r, db)
}
