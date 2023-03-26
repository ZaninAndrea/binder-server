package rest

import (
	"log"
	"math"
	"net/http"
	"time"

	"github.com/ZaninAndrea/binder-server/internal/mongo"
	"github.com/ZaninAndrea/binder-server/internal/mongo/op"
	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
)

func lerp(a, b, t float32) float32 {
	return a*(1-t) + b*t
}

func errorProbability(previousRepetition time.Time, time time.Time, halfLife float32) float32 {
	delay := float32(time.Sub(previousRepetition).Milliseconds())
	return float32(1 - math.Pow(2, float64(-delay/halfLife)))
}

func processRepetitions(repetitions []*mongo.Repetition) bson.M {
	slices.SortFunc(repetitions, func(a, b *mongo.Repetition) bool {
		if a == nil {
			return true
		} else if b == nil {
			return false
		}

		return a.Date.Before(b.Date)
	})

	var factor float32 = 0
	var halfLife float32 = 0
	var previousRepetition time.Time
	var correctRepetitions int = 0

	for i, repetition := range repetitions {
		if repetition.Quality >= 3 {
			correctRepetitions++
		}

		if i == 0 {
			factor = 2.5
			halfLife = 6.58 * 24 * 3600 * 1000
			previousRepetition = repetition.Date
			continue
		}

		usefulness := 10 * errorProbability(previousRepetition, repetition.Date, halfLife)
		if usefulness > 2 {
			usefulness = 2
		}

		quality := float32(repetition.Quality)
		factorUpdate := 0.1 - (5-quality)*(0.08+(5-quality)*0.02)
		factor = factor + factorUpdate*usefulness
		if factor < 1.3 {
			factor = 1.3
		} else if factor > 2.5 {
			factor = 2.5
		}

		if repetition.Quality < 3 {
			halfLife = halfLife / 2
		} else {
			halfLife = halfLife * lerp(1, factor*2, usefulness)
		}
	}

	update := bson.M{
		"cards.$.factor":             factor,
		"cards.$.halfLife":           halfLife,
		"cards.$.totalRepetitions":   len(repetitions),
		"cards.$.correctRepetitions": correctRepetitions,
	}
	if len(repetitions) > 0 {
		update["cards.$.lastRepetition"] = repetitions[len(repetitions)-1].Date
	}

	return update
}

func setupCardRoutes(r *gin.Engine, db *mongo.Database) {
	r.POST("/decks/:deckId/cards", Authenticated([]string{"user"}), func(c *gin.Context) {
		// Load user
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		// Load deck
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

		// Parse card
		var payload struct {
			Front string `json:"front"`
			Back  string `json:"back"`
		}
		err = c.ShouldBindJSON(&payload)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid payload")
			return
		}
		cardId := uuid.NewString()
		newCard := mongo.Card{
			ID:                 cardId,
			Front:              payload.Front,
			Back:               payload.Back,
			Factor:             2.5,
			LastRepetition:     nil,
			HalfLife:           0,
			TotalRepetitions:   0,
			CorrectRepetitions: 0,
			Paused:             false,
		}

		// Add card
		_, err = db.Decks.UpdateById(deck.ID, mongo.UpdateDocument{
			op.Push: bson.M{
				"cards": newCard,
			},
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to save card")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, cardId)
	})

	r.PUT("/decks/:deckId/cards/:cardId", Authenticated([]string{"user"}), func(c *gin.Context) {
		// Load user
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		// Load deck
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

		// Parse update
		var payload struct {
			Front  *string `json:"front"`
			Back   *string `json:"back"`
			Paused *bool   `json:"paused"`
		}
		err = c.ShouldBindJSON(&payload)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid payload")
			return
		}
		update := bson.M{}
		if payload.Front != nil {
			update["cards.$.front"] = payload.Front
		}
		if payload.Back != nil {
			update["cards.$.back"] = payload.Back
		}
		if payload.Paused != nil {
			update["cards.$.paused"] = payload.Paused
		}

		// Apply update
		_, err = db.Decks.UpdateOne(bson.M{
			"_id":      deck.ID,
			"cards.id": c.Param("cardId"),
		}, mongo.UpdateDocument{
			op.Set: update,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to update card")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, "")
	})

	r.DELETE("/decks/:deckId/cards/:cardId", Authenticated([]string{"user"}), func(c *gin.Context) {
		// Load user
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		// Load deck
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

		// Apply update
		_, err = db.Transaction(
			30*time.Second,
			func(db *mongo.Database, s mongo.SessionContext) (any, error) {
				_, err := db.Decks.UpdateById(deck.ID, mongo.UpdateDocument{
					op.Pull: bson.M{
						"cards": bson.M{
							"id": c.Param("cardId"),
						},
					},
				})
				if err != nil {
					return nil, err
				}

				_, err = db.Repetitions.DeleteMany(bson.M{
					"deckId": deck.ID,
					"cardId": c.Param("cardId"),
				})
				if err != nil {
					return nil, err
				}

				return nil, nil
			},
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to delete card")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, "")
	})

	r.POST("/decks/:deckId/cards/:cardId/repetition", Authenticated([]string{"user"}), func(c *gin.Context) {
		// Load user
		exists, err, user := GetAuthenticatedUser(c, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the user")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusUnauthorized, "The authentication token is associated with a non-existent user")
			return
		}

		// Load deck
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

		// Parse payload
		var payload struct {
			Date    time.Time `json:"date"`
			Quality int       `json:"quality"`
		}
		err = c.ShouldBindJSON(&payload)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid payload")
			return
		}
		cardId := c.Param("cardId")

		// Determine to which calendar day the repetition belongs
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			log.Println(err)
			location = time.UTC
		}
		dayOfRepetition := payload.Date.
			Add(-time.Hour * time.Duration(user.EndOfDay)).
			In(location).
			Format("2006-01-02")

		_, err = db.Transaction(
			30*time.Second,
			func(db *mongo.Database, s mongo.SessionContext) (any, error) {
				// Insert the new repetition
				_, err := db.Repetitions.InsertOne(&mongo.Repetition{
					CardId:  cardId,
					DeckID:  deck.ID,
					Date:    payload.Date,
					Quality: payload.Quality,
				})
				if err != nil {
					return nil, err
				}

				// Compute the half-life and factor based on the full
				// repetitions history
				repetitions := []*mongo.Repetition{}
				err = db.Repetitions.FindAll(bson.M{
					"cardId": cardId,
					"deckId": deck.ID,
				}, &repetitions)
				if err != nil {
					return nil, err
				}

				cardUpdate := processRepetitions(repetitions)

				// Update the card with the new half-life and factor
				_, err = db.Decks.UpdateOne(bson.M{
					"_id":      deck.ID,
					"cards.id": cardId,
				}, mongo.UpdateDocument{
					op.Set: cardUpdate,
				})
				if err != nil {
					return nil, err
				}

				// Initialize the user's daily repetitions count if needed
				repetitionsField := "statistics.dailyRepetitions." + dayOfRepetition
				_, err = db.Users.UpdateOne(bson.M{
					"_id": user.ID,
					repetitionsField: bson.M{
						string(op.Exists): false,
					},
				}, mongo.UpdateDocument{
					op.Set: {
						repetitionsField: 0,
					},
				})
				if err != nil {
					return nil, err
				}

				// Increment the user's daily repetitions count
				_, err = db.Users.UpdateOne(bson.M{
					"_id": user.ID,
				}, mongo.UpdateDocument{
					op.Inc: bson.M{
						repetitionsField: 1,
					},
				})
				if err != nil {
					return nil, err
				}

				return nil, nil
			},
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to record repetion")
			restLogger.Error(err)
			return
		}

		c.String(http.StatusOK, "")
	})
}
