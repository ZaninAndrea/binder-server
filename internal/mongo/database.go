package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type database interface {
	Disconnect() error
	MongoDatabase() *mongo.Database
	Context() context.Context
}

type basicDatabase struct {
	mongoClient   *mongo.Client
	mongoDatabase *mongo.Database
	rootContext   context.Context
}

func (db *basicDatabase) Context() context.Context {
	return db.rootContext
}

func (db *basicDatabase) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.mongoClient.Disconnect(ctx)
}

func (db *basicDatabase) Connect(mongoUri string, mongoDatabase string) {
	db.rootContext = context.Background()
	ctx, cancel := context.WithTimeout(db.rootContext, 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		panic(err)
	}
	cancel()

	db.mongoClient = client
	db.mongoDatabase = client.Database(mongoDatabase)
}

func (db *basicDatabase) MongoDatabase() *mongo.Database {
	return db.mongoDatabase
}

type Database struct {
	basicDatabase
	Users       Collection[*User]
	Decks       Collection[*Deck]
	Repetitions Collection[*Repetition]
}

func Connect(mongoUri string, mongoDatabase string) *Database {
	db := new(Database)
	db.basicDatabase.Connect(mongoUri, mongoDatabase)

	db.Users = NewCollection[*User](db, "users")
	db.Decks = NewCollection[*Deck](db, "decks")
	db.Repetitions = NewCollection[*Repetition](db, "repetitions")

	return db
}

func copy(ctx context.Context, db *Database) *Database {
	newDB := (*db)
	newDB.rootContext = ctx

	newDB.Users = NewCollection[*User](&newDB, "users")
	newDB.Decks = NewCollection[*Deck](&newDB, "decks")
	newDB.Repetitions = NewCollection[*Repetition](&newDB, "repetitions")

	return &newDB
}

type SessionContext mongo.SessionContext
type transactionFunction func(*Database, SessionContext) (interface{}, error)

func (db *Database) Transaction(timeout time.Duration, fn transactionFunction) (interface{}, error) {
	opts := options.Session()
	opts.SetDefaultReadConcern(readconcern.Majority())
	opts.SetDefaultWriteConcern(writeconcern.New(
		writeconcern.WMajority(),
		writeconcern.J(true),
		writeconcern.WTimeout(30*time.Second),
	))
	opts.SetCausalConsistency(true)

	session, err := db.mongoClient.StartSession(opts)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(db.rootContext, timeout)
	defer cancel()
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(s mongo.SessionContext) (interface{}, error) {
		sessionDB := copy(s, db)
		return fn(sessionDB, s)
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
