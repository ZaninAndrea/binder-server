package rest

import (
	"net/http"
	"time"

	"github.com/ZaninAndrea/binder-server/internal/mongo"
	"github.com/gin-gonic/gin"
	"github.com/nbutton23/zxcvbn-go"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword returns the bcrypt hash of the passed password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash checks whether the passed password is the same as the
// one stored in the hash (collision probability is negligible)
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func setupUserRoutes(r *gin.Engine, db *mongo.Database, jwtSecret []byte) {
	r.POST("/users", func(c *gin.Context) {
		// Parse request
		var payload struct {
			Email    string
			Password string
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.String(http.StatusBadRequest, "The payload is not a valid user JSON")
			restLogger.Error(err)
			return
		} else if payload.Email == "" {
			c.String(http.StatusBadRequest, "The email cannot be empty")
			return
		}

		passwordStrength := zxcvbn.PasswordStrength(payload.Password, []string{payload.Email})
		if passwordStrength.Score < 2 {
			c.String(http.StatusBadRequest, "The password is too weak")
			return
		}

		password, err := HashPassword(payload.Password)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create a new user")
			restLogger.Error(err)
			return
		}

		// Check that no user with the same email exists
		var otherUser mongo.User
		exists, err := db.Users.FindOneIfExists(bson.M{
			"email": payload.Email,
		}, &otherUser)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create a new user")
			restLogger.Error(err)
			return
		} else if exists {
			c.String(http.StatusBadRequest, "A user with this email already exists")
			return
		}

		// Add user in the database
		id, err := db.Users.InsertOne(&mongo.User{
			Email:    payload.Email,
			Password: password,
			Plan:     mongo.BasicPlan,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create a new user")
			restLogger.Error(err)
			return
		}

		// Create the JWT for the user
		token, err := SignToken("user", id.Hex(), jwtSecret)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create authentication token")
			restLogger.Error(err)
			return
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"id":    id.Hex(),
			"token": token,
		})
	})

	r.POST("/users/login", func(c *gin.Context) {
		// Parse request
		var payload struct {
			Email    string
			Password string
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.String(http.StatusBadRequest, "The payload is not a valid user JSON")
			restLogger.Error(err)
			return
		} else if payload.Email == "" {
			c.String(http.StatusBadRequest, "The email cannot be empty")
			return
		}

		// Load the user from the database
		var user mongo.User
		exists, err := db.Users.FindOneIfExists(bson.M{
			"email": payload.Email,
		}, &user)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusBadRequest, "A user with this email does not exist")
			return
		}

		// Check that the password is correct
		correct := CheckPasswordHash(payload.Password, user.Password)
		if !correct {
			c.String(http.StatusBadRequest, "Wrong password")
			return
		}

		// Create the JWT for the user
		token, err := SignToken("user", user.ID.Hex(), jwtSecret)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create authentication token")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, token)
	})

	r.GET("/users", Authenticated([]string{"user"}), func(c *gin.Context) {
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusBadRequest, "The specified user does not exist")
			return
		}

		c.JSON(200, user)
	})

	r.DELETE("/users", Authenticated([]string{"user"}), func(c *gin.Context) {
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusBadRequest, "The specified user does not exist")
			return
		}

		// Delete the user and the associated resources
		_, err = db.Transaction(
			60*time.Second,
			func(db *mongo.Database, s mongo.SessionContext) (any, error) {
				// Delete the deck and the associated repetitions
				decks := []*mongo.Deck{}
				err := db.Decks.FindAll(bson.M{
					"owner": user.ID,
				}, &decks)
				if err != nil {
					return nil, err
				}

				for _, deck := range decks {
					_, err := db.Decks.DeleteById(deck.ID)
					if err != nil {
						return nil, err
					}

					_, err = db.Repetitions.DeleteMany(bson.M{
						"deckId": deck.ID,
					})
					if err != nil {
						return nil, err
					}
				}

				db.Users.DeleteById(user.ID)

				return nil, nil
			},
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to delete the user")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, "")
	})
}
