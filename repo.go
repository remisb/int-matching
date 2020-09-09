package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// MongoTimeout has default timeout value for mongo db connection.
const mongoTimeout = 150 * time.Second

var EmptyFilter = bson.M{}

type Repo struct {
	mngClient *mongo.Client
	DbName    string
}

func NewRepo(mngClient *mongo.Client, dbName string) *Repo {
	return &Repo{mngClient, dbName}
}

func (r *Repo) GetMatching(ctx context.Context, id primitive.ObjectID) (Matching, error) {
	filter := bson.M{"_id": id}
	return r.readMatching(ctx, filter)
}

// getMatchingBySummaryId returns a list of matchings for passed summaryId.
func (r *Repo) GetMatchingBySummaryId(ctx context.Context, summaryID primitive.ObjectID) (Matching, error) {
	filter := bson.M{"summaryId": summaryID}
	return r.readMatching(ctx, filter)
}

// GetAllMatchings retrieves a list of matchings from the database.
func (r *Repo) GetAllMatchings(ctx context.Context) ([]*Matching, error) {
	return r.readMatchings(ctx, EmptyFilter)
}

func (r Repo) readMatching(ctx context.Context, filter interface{}) (Matching, error) {
	cursor, _ := r.getMatchingCollection().Find(ctx, filter)

	matching := Matching{}
	if cursor.Next(ctx) {
		err := cursor.Decode(&matching)
		if err != nil {
			return matching, err
		}
	}
	return matching, nil
}

func (r Repo) readMatchings(ctx context.Context, filter interface{}) ([]*Matching, error) {
	result := make([]*Matching, 0)
	cursor, _ := r.getMatchingCollection().Find(ctx, filter)
	for cursor.Next(ctx) {
		matching := Matching{}
		err := cursor.Decode(&matching)
		if err != nil {
			return result, err
		}
		result = append(result, &matching)
	}
	return result, nil
}

func (r *Repo) getDb() *mongo.Database {
	return r.mngClient.Database(r.DbName)
}

func (r *Repo) getMatchingCollection() *mongo.Collection {
	return r.getDb().Collection("matching")
}

func (r *Repo) getSummaryCollection() *mongo.Collection {
	return r.getDb().Collection("summary")
}

func (r *Repo) CreateMatching(ctx context.Context, matching Matching) (*mongo.InsertOneResult, error) {
	insertResult, err := r.saveNewMatching(ctx, matching)
	if err != nil {
		return insertResult, err
	}
	return insertResult, nil
}

func (r *Repo) UpdateMatching(ctx context.Context, matching Matching) (int64, error) {
	if primitive.NilObjectID == matching.Id {
		// add new matching
		_, err := r.saveNewMatching(ctx, matching)
		if err != nil {
			return 0, err
		}
		return 1, nil
	}

	// update existing matching
	updateResult, err := r.updateMatching(ctx, matching)
	if err != nil {
		return 0, err
	}

	return updateResult.ModifiedCount, nil
}

func (r *Repo) 	saveNewMatching(ctx context.Context, matching Matching) (*mongo.InsertOneResult, error){
	insert := bson.M{
		"summaryId": matching.SummaryId,
		"matchedSummaryId": matching.MatchedSummaryId,
		"matchRate":   matching.MatchRate,
		"createdAt": matching.CreatedAt,
	}
	return r.getMatchingCollection().InsertOne(ctx, insert)
}

func (r *Repo) updateMatching(ctx context.Context, matching Matching) (*mongo.UpdateResult, error) {
	filter := bson.M{"_id": matching.Id}
	update := bson.M{
		"$set": bson.M{
			"_id": matching.Id,
			"summaryId": matching.SummaryId,
			"matchedSummaryId": matching.MatchedSummaryId,
			"matchRate":   matching.MatchRate,
			"createdAt": matching.CreatedAt,
		},
	}

	return r.getMatchingCollection().UpdateOne(ctx, filter, update)
}

func AddTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, mongoTimeout)
}
