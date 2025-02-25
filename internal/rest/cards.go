package rest

import (
	"log"
	"math"
	"net/http"
	"time"

	"github.com/ZaninAndrea/binder-server/internal/achievements"
	"github.com/ZaninAndrea/binder-server/internal/mongo"
	"github.com/ZaninAndrea/binder-server/internal/mongo/op"
	"github.com/ZaninAndrea/binder-server/storage"
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
		if usefulness > 1.5 {
			usefulness = 1.5
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
			halfLife = halfLife * lerp(1, factor, usefulness)
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

func setupCardRoutes(r *gin.Engine, db *mongo.Database, storage *storage.BlobStorage) {
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
		front, err := ReplaceBase64ImagesWithFileLinks(payload.Front, storage)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to replace base64 images with file links")
			restLogger.Error(err)
			return
		}
		back, err := ReplaceBase64ImagesWithFileLinks(payload.Back, storage)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to replace base64 images with file links")
			restLogger.Error(err)
			return
		}
		newCard := mongo.Card{
			ID:                 cardId,
			Front:              front,
			Back:               back,
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
			front, err := ReplaceBase64ImagesWithFileLinks(*payload.Front, storage)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to replace base64 images with file links")
				restLogger.Error(err)
				return
			}
			update["cards.$.front"] = front
		}
		if payload.Back != nil {
			back, err := ReplaceBase64ImagesWithFileLinks(*payload.Back, storage)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to replace base64 images with file links")
				restLogger.Error(err)
				return
			}
			update["cards.$.back"] = back
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

	r.PUT("/decks/:deckId/cards/:cardId/move", Authenticated([]string{"user"}), func(c *gin.Context) {
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

		// Load target deck
		var payload struct {
			NewDeckId string `json:"newDeckId"`
		}
		err = c.ShouldBindJSON(&payload)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid payload")
			return
		}

		newDeckId, err := primitive.ObjectIDFromHex(payload.NewDeckId)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid new deck id")
			return
		}

		var newDeck mongo.Deck
		exists, err = db.Decks.FindByIdIfExists(newDeckId, &newDeck)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the new deck")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusBadRequest, "The specified new deck does not exist")
			return
		} else if newDeck.Owner != user.ID {
			c.String(http.StatusUnauthorized, "You are not the owner of the new deck")
			return
		}

		cardId := c.Param("cardId")
		// Apply update
		newCard, err := db.Transaction(
			30*time.Second,
			func(db *mongo.Database, s mongo.SessionContext) (any, error) {
				_, err := db.Decks.UpdateById(deck.ID, mongo.UpdateDocument{
					op.Pull: bson.M{
						"cards": bson.M{
							"id": cardId,
						},
					},
				})
				if err != nil {
					return nil, err
				}

				var card mongo.Card
				for _, c := range deck.Cards {
					if c.ID == cardId {
						card = c
						break
					}
				}

				_, err = db.Decks.UpdateById(newDeck.ID, mongo.UpdateDocument{
					op.Push: bson.M{
						"cards": card,
					},
				})
				if err != nil {
					return nil, err
				}

				_, err = db.Repetitions.UpdateMany(bson.M{
					"deckId": deck.ID,
					"cardId": cardId,
				}, mongo.UpdateDocument{
					op.Set: bson.M{
						"deckId": newDeck.ID,
					},
				})
				if err != nil {
					return nil, err
				}

				return card, nil
			},
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to copy card")
			restLogger.Error(err)
			return
		}

		c.JSON(http.StatusOK, newCard)
	})

	r.PUT("/decks/:deckId/cards/:cardId/copy", Authenticated([]string{"user"}), func(c *gin.Context) {
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

		// Load target deck
		var payload struct {
			NewDeckId string `json:"newDeckId"`
		}
		err = c.ShouldBindJSON(&payload)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid payload")
			return
		}

		newDeckId, err := primitive.ObjectIDFromHex(payload.NewDeckId)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid new deck id")
			return
		}

		var newDeck mongo.Deck
		exists, err = db.Decks.FindByIdIfExists(newDeckId, &newDeck)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load the new deck")
			restLogger.Error(err)
			return
		} else if !exists {
			c.String(http.StatusBadRequest, "The specified new deck does not exist")
			return
		} else if newDeck.Owner != user.ID {
			c.String(http.StatusUnauthorized, "You are not the owner of the new deck")
			return
		}

		cardId := c.Param("cardId")
		// Apply update
		card, err := db.Transaction(
			30*time.Second,
			func(db *mongo.Database, s mongo.SessionContext) (any, error) {
				var card mongo.Card
				for _, c := range deck.Cards {
					if c.ID == cardId {
						card = c
						break
					}
				}

				card.ID = uuid.NewString()

				// Copy card
				_, err = db.Decks.UpdateById(newDeck.ID, mongo.UpdateDocument{
					op.Push: bson.M{
						"cards": card,
					},
				})
				if err != nil {
					return nil, err
				}

				// Copy repetitions
				repetitions := []*mongo.Repetition{}
				err = db.Repetitions.FindAll(bson.M{
					"deckId": deck.ID,
					"cardId": cardId,
				}, &repetitions)
				if err != nil {
					return nil, err
				}

				if len(repetitions) > 0 {
					for _, repetition := range repetitions {
						repetition.ID = primitive.NilObjectID
						repetition.DeckID = newDeck.ID
						repetition.CardId = card.ID
						_, err = db.Repetitions.InsertOne(repetition)
						if err != nil {
							return nil, err
						}
					}

					err = db.Repetitions.InsertMany(repetitions)
					if err != nil {
						return nil, err
					}
				}

				return card, nil
			},
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to delete card")
			restLogger.Error(err)
			return
		}

		c.JSON(http.StatusOK, card)
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

		updates, err := db.Transaction(
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

				var updatedUser mongo.User
				err = db.Users.FindById(user.ID, &updatedUser)
				if err != nil {
					return nil, err
				}

				// Update the user's achievements if needed
				updates := achievements.UpdateAchievements(&updatedUser)
				if len(updates) > 0 {
					mongoUpdates := bson.M{}
					for _, update := range updates {
						mongoUpdates["achievements."+update.ID] = update.Level
					}

					_, err = db.Users.UpdateById(user.ID, mongo.UpdateDocument{
						op.Set: mongoUpdates,
					})

					if err != nil {
						return nil, err
					}
				}

				return updates, nil
			},
		)

		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to record repetition")
			restLogger.Error(err)
			return
		}

		c.JSON(http.StatusOK, updates)
	})
}
