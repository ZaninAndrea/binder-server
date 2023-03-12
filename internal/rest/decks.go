package rest

import (
	"net/http"
	"time"

	"github.com/ZaninAndrea/binder-server/internal/mongo"
	"github.com/ZaninAndrea/binder-server/internal/mongo/op"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupDeckRoutes(r *gin.Engine, db *mongo.Database) {
	r.POST("/decks", Authenticated([]string{"user"}), func(c *gin.Context) {
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		var payload struct {
			Name string
		}
		err = c.ShouldBindJSON(&payload)
		if err != nil {
			c.String(http.StatusBadRequest, "The payload is invalid")
			return
		} else if payload.Name == "" {
			c.String(http.StatusBadRequest, "You must specify an non-empty `name` field")
			return
		}

		deckId, err := db.Decks.InsertOne(&mongo.Deck{
			Name:     payload.Name,
			Archived: false,
			Cards:    []mongo.Card{},
			Owner:    user.ID,
		})

		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create the deck")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, deckId.Hex())
	})

	r.GET("/decks", Authenticated([]string{"user"}), func(c *gin.Context) {
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		decks := []*mongo.Deck{}
		err = db.Decks.FindAll(bson.M{
			"owner": user.ID,
		}, &decks)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the decks")
			restLogger.Error(err)
			return
		}

		c.JSON(http.StatusOK, decks)
	})

	r.GET("/decks/:deckId", Authenticated([]string{"user"}), func(c *gin.Context) {
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		rawId := c.Param("deckId")
		deckId, err := primitive.ObjectIDFromHex(rawId)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid deck id")
			return
		}

		var deck mongo.Deck
		exists, err = db.Decks.FindByIdIfExists(deckId, &deck)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the deck")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusBadRequest, "The specified deck does not exist")
			return
		} else if deck.Owner != user.ID {
			c.String(http.StatusUnauthorized, "You are not the owner of this deck")
			return
		}

		c.JSON(http.StatusOK, deck)
	})

	r.PUT("/decks/:deckId", Authenticated([]string{"user"}), func(c *gin.Context) {
		// Check that the authenticated user exists
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		// Parse query parameters
		rawId := c.Param("deckId")
		deckId, err := primitive.ObjectIDFromHex(rawId)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid deck id")
			return
		}

		// Parse the payload
		var query struct {
			Archived *bool   `json:"archived"`
			Name     *string `json:"name"`
		}
		err = c.ShouldBindJSON(&query)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid body: %s", err.Error())
			return
		}

		// Check that the authenticated user is the deck's owner
		var deck mongo.Deck
		exists, err = db.Decks.FindByIdIfExists(deckId, &deck)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the deck")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusBadRequest, "The specified deck does not exist")
			return
		} else if deck.Owner != user.ID {
			c.String(http.StatusUnauthorized, "You are not the owner of this deck")
			return
		}

		// Update the deck
		update := bson.M{}
		if query.Archived != nil {
			update["archived"] = *query.Archived
		}
		if query.Name != nil {
			update["name"] = *query.Name
		}
		_, err = db.Decks.UpdateById(deck.ID, mongo.UpdateDocument{
			op.Set: update,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to update the deck")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, "")
	})

	r.DELETE("/decks/:deckId", Authenticated([]string{"user"}), func(c *gin.Context) {
		// Check that the authenticated user exists
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		// Parse query parameters
		rawId := c.Param("deckId")
		deckId, err := primitive.ObjectIDFromHex(rawId)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid deck id")
			return
		}

		// Check that the authenticated user is the deck's owner
		var deck mongo.Deck
		exists, err = db.Decks.FindByIdIfExists(deckId, &deck)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the deck")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusBadRequest, "The specified deck does not exist")
			return
		} else if deck.Owner != user.ID {
			c.String(http.StatusUnauthorized, "You are not the owner of this deck")
			return
		}

		// Delete the deck and the associated repetitions
		_, err = db.Transaction(
			30*time.Second,
			func(db *mongo.Database, s mongo.SessionContext) (any, error) {
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

				return nil, nil
			},
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to delete the deck")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, "")
	})

}
