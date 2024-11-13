package main

import (
	"fmt"
	"os"

	"github.com/ZaninAndrea/binder-server/internal/log"
	"github.com/ZaninAndrea/binder-server/internal/mongo"
	"github.com/ZaninAndrea/binder-server/internal/mongo/op"
	"github.com/ZaninAndrea/binder-server/internal/rest"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
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

	// Fetch all the decks
	var decks []*mongo.Deck
	err = db.Decks.FindAll(bson.M{}, &decks)
	if err != nil {
		mainLogger.Fatal("Error fetching decks: ", err)
	}

	// Migrate all the decks
	for i, deck := range decks {
		fmt.Println("Migrating deck ", i+1, "/", len(decks))

		// Migrate each card
		for _, card := range deck.Cards {
			repetitions := []*mongo.Repetition{}
			err = db.Repetitions.FindAll(bson.M{
				"cardId": card.ID,
				"deckId": deck.ID,
			}, &repetitions)
			if err != nil {
				mainLogger.Fatal("Error fetching repetitions: ", err)
			}

			// Update the card scheduling data
			cardUpdate := rest.ProcessRepetitions(repetitions)

			// Update the card with the new half-life and factor
			_, err = db.Decks.UpdateOne(bson.M{
				"_id":      deck.ID,
				"cards.id": card.ID,
			}, mongo.UpdateDocument{
				op.Set: cardUpdate,
			})
			if err != nil {
				mainLogger.Fatal("Error updating card: ", err)
			}
		}
	}
}
