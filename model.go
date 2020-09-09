package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Matching struct {
	Id               primitive.ObjectID `json:"id" bson:"_id"`
	SummaryId        primitive.ObjectID `json:"summaryId" bson:"summaryId"`
	MatchedSummaryId primitive.ObjectID `json:"matchedSummaryId" bson:"matchedSummaryId"`
	MatchRate        int                `json:"matchRate" bson:"matchRate"`
	CreatedAt        time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}
