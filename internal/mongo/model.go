package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Model interface {
	SetID(primitive.ObjectID)
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

type BasicModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

func (c *BasicModel) SetCreatedAt(timestamp time.Time) {
	c.CreatedAt = timestamp
}

func (c *BasicModel) SetUpdatedAt(timestamp time.Time) {
	c.UpdatedAt = timestamp
}

func (c *BasicModel) SetID(id primitive.ObjectID) {
	c.ID = id
}

type UserPlan string

const (
	VIPPlan   UserPlan = "VIP"
	BasicPlan UserPlan = "BASIC"
)

type User struct {
	BasicModel `bson:",inline"`
	Email      string   `bson:"email" json:"email"`
	Password   string   `bson:"password" json:"-"`
	Plan       UserPlan `bson:"plan" json:"plan"`
}

type Deck struct {
	BasicModel `bson:",inline"`
	Archived   bool               `bson:"archived" json:"archived"`
	Name       string             `bson:"name" json:"name"`
	Cards      []Card             `bson:"cards" json:"cards"`
	Owner      primitive.ObjectID `bson:"owner" json:"-"`
}

type Card struct {
	ID                 string     `bson:"id" json:"id"`
	Front              string     `bson:"front" json:"front"`
	Back               string     `bson:"back" json:"back"`
	Factor             float32    `bson:"factor" json:"factor"`
	HalfLife           float32    `bson:"halfLife" json:"halfLife"`
	TotalRepetitions   float32    `bson:"totalRepetitions" json:"totalRepetitions"`
	CorrectRepetitions float32    `bson:"correctRepetitions" json:"correctRepetitions"`
	LastRepetition     *time.Time `bson:"lastRepetition" json:"lastRepetition"`
	Paused             bool       `bson:"paused" json:"paused"`
}

type Repetition struct {
	BasicModel `bson:",inline"`
	CardId     string             `bson:"cardId" json:"cardId"`
	DeckID     primitive.ObjectID `bson:"deckId" json:"deckId"`
	Date       time.Time          `bson:"date" json:"date"`
	Quality    int                `bson:"quality" json:"quality"`
}
