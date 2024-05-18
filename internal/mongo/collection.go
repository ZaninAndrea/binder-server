package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	op "github.com/ZaninAndrea/binder-server/internal/mongo/op"
)

var ErrNoDocuments error = mongo.ErrNoDocuments

type Collection[Record Model] struct {
	collection *mongo.Collection
	db         database
}

func NewCollection[Record Model](db database, collectionName string) Collection[Record] {
	return Collection[Record]{
		collection: db.MongoDatabase().Collection(collectionName),
		db:         db,
	}
}

func (c *Collection[Record]) GetTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.db.Context(), 10*time.Second)
}

func (c *Collection[Record]) Collection() *mongo.Collection {
	return c.collection
}

func (c *Collection[Record]) InsertOne(document Record) (primitive.ObjectID, error) {
	ctx, cancel := c.GetTimeoutContext()

	document.SetID(primitive.NilObjectID)
	document.SetCreatedAt(time.Now())
	document.SetUpdatedAt(time.Now())

	res, err := c.collection.InsertOne(ctx, document)
	cancel()

	if err != nil {
		return primitive.ObjectID{}, err
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

func (c *Collection[Record]) InsertMany(documents []Record) error {
	ctx, cancel := c.GetTimeoutContext()

	docs := make([]interface{}, len(documents))
	for i := range documents {
		documents[i].SetID(primitive.NilObjectID)
		documents[i].SetCreatedAt(time.Now())
		documents[i].SetUpdatedAt(time.Now())
		docs[i] = documents[i]
	}

	_, err := c.collection.InsertMany(ctx, docs)
	cancel()

	if err != nil {
		return err
	}

	return nil
}

func (c *Collection[Record]) InsertOneIfNotExists(filter interface{}, document Record) (primitive.ObjectID, bool, error) {
	ctx, cancel := c.GetTimeoutContext()

	document.SetID(primitive.NilObjectID)
	document.SetCreatedAt(time.Now())
	document.SetUpdatedAt(time.Now())

	opts := options.Update()
	opts.SetUpsert(true)
	res, err := c.collection.UpdateOne(ctx, filter, bson.M{"$setOnInsert": document}, opts)
	cancel()
	if err != nil {
		return primitive.ObjectID{}, false, err
	}

	if res.MatchedCount > 0 {
		return primitive.ObjectID{}, true, nil
	}

	return res.UpsertedID.(primitive.ObjectID), false, nil
}

func (c *Collection[Record]) FindOne(filter interface{}, result Record) error {
	ctx, cancel := c.GetTimeoutContext()

	res := c.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return res.Err()
	}
	cancel()

	return res.Decode(result)
}

func (c *Collection[Record]) FindById(id primitive.ObjectID, result Record) error {
	return c.FindOne(bson.M{
		"_id": id,
	}, result)
}

func (c *Collection[Record]) FindByIdIfExists(id primitive.ObjectID, result Record) (bool, error) {
	return c.FindOneIfExists(bson.M{
		"_id": id,
	}, result)
}

func (c *Collection[Record]) FindOneIfExists(filter interface{}, result Record) (bool, error) {
	ctx, cancel := c.GetTimeoutContext()
	res := c.collection.FindOne(ctx, filter)
	cancel()

	if res.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else if res.Err() != nil {
		return false, res.Err()
	}

	err := res.Decode(result)
	return true, err
}

func (c *Collection[Record]) Exists(filter interface{}) (bool, error) {
	ctx, cancel := c.GetTimeoutContext()

	count, err := c.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	cancel()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (c *Collection[Record]) Count(filter interface{}) (int64, error) {
	ctx, cancel := c.GetTimeoutContext()

	total, err := c.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	cancel()

	return total, nil
}

func (c *Collection[Record]) FindAll(filter interface{}, results *[]Record) error {
	ctx, cancel := c.GetTimeoutContext()
	defer cancel()

	cur, err := c.collection.Find(ctx, filter)
	if err != nil {
		return err
	}

	return cur.All(ctx, results)
}

type UpdateDocument map[op.Operator]bson.M

func (c *Collection[Record]) UpdateOne(filter interface{}, update UpdateDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	ctx, cancel := c.GetTimeoutContext()

	// Modify the update document to ensure that the createdAt
	// and updatedAt fields are updated correctly
	if set, ok := update[op.Set]; ok {
		if _, ok := set["updatedAt"]; !ok {
			set["updatedAt"] = time.Now()
		}
	} else if setOnInsert, ok := update[op.SetOnInsert]; ok {
		if _, ok := setOnInsert["createdAt"]; !ok {
			setOnInsert["createdAt"] = time.Now()
		}
		if _, ok := setOnInsert["updatedAt"]; !ok {
			setOnInsert["updatedAt"] = time.Now()
		}
	} else {
		update[op.Set] = bson.M{
			"updatedAt": time.Now(),
		}
	}

	res, err := c.collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, err
	}
	cancel()

	return res, nil
}

func (c *Collection[Record]) UpdateMany(filter interface{}, update UpdateDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	ctx, cancel := c.GetTimeoutContext()

	// Modify the update document to ensure that the createdAt
	// and updatedAt fields are updated correctly
	if set, ok := update[op.Set]; ok {
		if _, ok := set["updatedAt"]; !ok {
			set["updatedAt"] = time.Now()
		}
	} else if setOnInsert, ok := update[op.SetOnInsert]; ok {
		if _, ok := setOnInsert["createdAt"]; !ok {
			setOnInsert["createdAt"] = time.Now()
		}
		if _, ok := setOnInsert["updatedAt"]; !ok {
			setOnInsert["updatedAt"] = time.Now()
		}
	} else {
		update[op.Set] = bson.M{
			"updatedAt": time.Now(),
		}
	}

	res, err := c.collection.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, err
	}
	cancel()

	return res, nil
}

func (c *Collection[Record]) UpdateById(id primitive.ObjectID, update UpdateDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.UpdateOne(bson.M{"_id": id}, update, opts...)
}

func (c *Collection[Record]) DeleteById(id primitive.ObjectID) (int64, error) {
	ctx, cancel := c.GetTimeoutContext()

	res, err := c.collection.DeleteOne(ctx, bson.M{
		"_id": id,
	})
	if err != nil {
		return 0, err
	}
	cancel()

	return res.DeletedCount, nil
}

func (c *Collection[Record]) DeleteMany(filter bson.M) (int64, error) {
	ctx, cancel := c.GetTimeoutContext()

	res, err := c.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	cancel()

	return res.DeletedCount, nil
}
