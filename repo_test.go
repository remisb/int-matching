package main

import (
	"context"
	iss "github.com/matryer/is"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	cfg      Config
	dbClient *mongo.Client
	repo     *Repo
)

const (
	summaryId1 = "5e458de13f2d3aad1bf0bb6f"
	summaryId2 = "5e458de13f2d3aad1bf0bb70"
)

type Summary struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	ProfileId primitive.ObjectID `json:"profileId" bson:"profileId"`
}

func TestAddUpdateDeleteMatching(t *testing.T) {
	t.Run("Remove matchings", testRemoveMatchings)
	t.Run("Remove summaries", testRemoveSummaries)
	t.Run("Get all matchings is empty", getGetMatchingCount(0))
	t.Run("Init test summaries", testInitTestSummaries)
	t.Run("Add new matchings", testAddMatching)
	t.Run("Update matching", testUpdateMatching)
	t.Run("Get all matchings returns 2", getGetMatchingCount(2))
	t.Run("clean database", testCleanDatabase)
}

func testInitTestSummaries(t *testing.T) {
	tests := []struct {
		name string
		args Summary
	}{
		{
			name: "add summary1",
			args: Summary{
				Id:        makeObjectId(t, summaryId1),
				ProfileId: makeObjectId(t, "5e458def3f2d3aad1bf0bb86"),
			},
		}, {
			name: "add summary2",
			args: Summary{
				Id:        makeObjectId(t, "5e458de13f2d3aad1bf0bb70"),
				ProfileId: makeObjectId(t, "5e458def3f2d3aad1bf0bb86"),
			},
		},
	}

	ctx := context.TODO()
	is := iss.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.getSummaryCollection().InsertOne(ctx, tt.args)
			is.NoErr(err)
			is.True(result.InsertedID.(primitive.ObjectID) != primitive.NilObjectID)
		})
	}
}

func testCleanDatabase(t *testing.T) {
	testRemoveMatchings(t)
	testRemoveSummaries(t)
}

func testRemoveMatchings(t *testing.T) {
	_, err := repo.getMatchingCollection().DeleteMany(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
}

func testRemoveSummaries(t *testing.T) {
	_, err := repo.getSummaryCollection().DeleteMany(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
}

func testUpdateMatching(t *testing.T) {
	ctx := context.Background()
	matching, err := repo.GetMatchingBySummaryId(ctx, makeObjectId(t, "5e458de13f2d3aad1bf0bb6f"))
	is := iss.New(t)
	is.NoErr(err)
	newMatchRate := matching.MatchRate + 10
	matching.MatchRate = newMatchRate
	updateCount, err := repo.UpdateMatching(ctx, matching)
	is.NoErr(err)
	is.Equal(updateCount, int64(1))

	retrievedMatching, err := repo.GetMatching(ctx, matching.Id)
	is.NoErr(err)
	is.Equal(retrievedMatching.Id, matching.Id)
	is.Equal(retrievedMatching.SummaryId, matching.SummaryId)
	is.Equal(retrievedMatching.MatchedSummaryId, matching.MatchedSummaryId)
	is.Equal(retrievedMatching.CreatedAt, matching.CreatedAt)
	is.Equal(retrievedMatching.MatchRate, newMatchRate)
}

func testAddMatching(t *testing.T) {
	match1 := Matching{
		SummaryId:        makeObjectId(t, summaryId1),
		MatchedSummaryId: makeObjectId(t, summaryId2),
		MatchRate:        15,
		CreatedAt:        time.Now().Truncate(time.Millisecond),
	}

	match2 := Matching{
		SummaryId:        makeObjectId(t, summaryId2),
		MatchedSummaryId: makeObjectId(t, summaryId1),
		MatchRate:        25,
		CreatedAt:        time.Now().Truncate(time.Millisecond),
	}

	tests := []struct {
		name         string
		matching     Matching
		wantMatching Matching
	}{
		{
			name:         "add match for summary summaryId1",
			matching:     match1,
			wantMatching: match1,
		},
		{
			name:          "add match for summary summaryId2",
			matching:     match2,
			wantMatching: match2,
		},
	}
	ctx := context.Background()
	is := iss.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			insRes, err := repo.CreateMatching(ctx, tt.matching)
			is.NoErr(err)
			is.True(insRes.InsertedID != primitive.NilObjectID)

			//matching, err := repo.GetMatching(ctx, insRes.InsertedID.(primitive.ObjectID))
			tt.wantMatching.Id = insRes.InsertedID.(primitive.ObjectID)
			matching, err := repo.GetMatching(ctx, tt.wantMatching.Id)
			is.NoErr(err)
			is.Equal(matching.Id, tt.wantMatching.Id)
			is.Equal(matching.SummaryId, tt.wantMatching.SummaryId)
			is.Equal(matching.MatchedSummaryId, tt.wantMatching.MatchedSummaryId)
			is.Equal(matching.MatchRate, tt.wantMatching.MatchRate)
			is.True(matching.CreatedAt.Equal(tt.wantMatching.CreatedAt))
		})
	}

}

func TestNewRepo(t *testing.T) {
	newRepo := NewRepo(client, mongoDbConfig.DbName)
	is := iss.New(t)
	is.True(newRepo != nil)
	is.Equal(newRepo.DbName, mongoDbConfig.DbName)
	is.Equal(newRepo.mngClient, client)
}

func TestMain(m *testing.M) {
	setUp()
	exitVal := m.Run()
	tearDown()
	os.Exit(exitVal)
}

func initConfig() {
	cfg = Config{
		DriverName: "mongodb",
		DbHost:     "localhost", //# aws winawin mongodb public ip
		DbPort:     "27017",
		DbName:     "winawin_test_2",
	}
}

func setUp() {
	log.Print("Setup")
	dbClient = initDbForTest()
	repo = NewRepo(dbClient, cfg.DbName)
}

func tearDown() {
	// init data
	log.Print("TearDown")
	delResult, err := repo.getMatchingCollection().DeleteMany(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("On tearDown - delResult.deleteCount:", delResult.DeletedCount)
}

func initDbForTest() *mongo.Client {
	initConfig()

	mongoClient, err := NewMongoClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return mongoClient
}

func testGetAllMatchingsCount(t *testing.T, count int) {
	is := iss.New(t)
	matchings, err := repo.GetAllMatchings(context.TODO())
	is.NoErr(err)
	is.True(len(matchings) == count)
}

func getGetMatchingCount(count int) func(t *testing.T) {
	return func(t *testing.T) {
		testGetAllMatchingsCount(t, count)
	}
}

func TestRepo_GetMatching(t *testing.T) {
	type fields struct {
		mngClient *mongo.Client
		DbName    string
	}
	type args struct {
		ctx context.Context
		id  primitive.ObjectID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Matching
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repo{
				mngClient: tt.fields.mngClient,
				DbName:    tt.fields.DbName,
			}
			got, err := r.GetMatching(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMatching() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMatching() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//func TestRepo_GetMatchingBySummaryId(t *testing.T) {
//	tests := []struct {
//		name          string
//		summaryId     primitive.ObjectID
//		wantSummaryId primitive.ObjectID
//	}{
//		{
//			name:          "get summary1 by summaryId",
//			summaryId:     makeObjectId(t, summaryId1),
//			wantSummaryId: makeObjectId(t, summaryId1),
//		},
//	}
//	ctx := context.Background()
//	is := iss.New(t)
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := repo.GetMatchingBySummaryId(ctx, tt.summaryId)
//			is.NoErr(err)
//			is.Equal(got.SummaryId, tt.wantSummaryId)
//		})
//	}
//}

func makeObjectId(t *testing.T, id string) primitive.ObjectID {
	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		t.Fatal(err)
	}
	return oId
}

func TestRepo_UpdateMatching(t *testing.T) {

	t.Log("UpdateMatching tests started.")
	type args struct {
		matching Matching
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Add new Matching No 1",
			args: args{Matching{
				Id:               primitive.NilObjectID,
				CreatedAt:        time.Now(),
				SummaryId:        makeObjectId(t, "5e458de13f2d3aad1bf0bb71"),
				MatchedSummaryId: makeObjectId(t, "5e458de13f2d3aad1bf0bb6f"),
				MatchRate:        15,
			}},
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.UpdateMatching(context.Background(), tt.args.matching)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMatching() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UpdateMatching() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_getMatchingCollection(t *testing.T) {
	type fields struct {
		mngClient *mongo.Client
		DbName    string
	}
	tests := []struct {
		name   string
		fields fields
		want   *mongo.Collection
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repo{
				mngClient: tt.fields.mngClient,
				DbName:    tt.fields.DbName,
			}
			if got := r.getMatchingCollection(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMatchingCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_readMatching(t *testing.T) {
	type fields struct {
		mngClient *mongo.Client
		DbName    string
	}
	type args struct {
		ctx    context.Context
		filter interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Matching
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Repo{
				mngClient: tt.fields.mngClient,
				DbName:    tt.fields.DbName,
			}
			got, err := r.readMatching(tt.args.ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMatching() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readMatching() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_readMatchings(t *testing.T) {
	type fields struct {
		mngClient *mongo.Client
		DbName    string
	}
	type args struct {
		ctx    context.Context
		filter interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Matching
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Repo{
				mngClient: tt.fields.mngClient,
				DbName:    tt.fields.DbName,
			}
			got, err := r.readMatchings(tt.args.ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMatchings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readMatchings() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_saveNewMatching(t *testing.T) {
	type fields struct {
		mngClient *mongo.Client
		DbName    string
	}
	type args struct {
		ctx      context.Context
		matching Matching
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *mongo.InsertOneResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repo{
				mngClient: tt.fields.mngClient,
				DbName:    tt.fields.DbName,
			}
			got, err := r.saveNewMatching(tt.args.ctx, tt.args.matching)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveNewMatching() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("saveNewMatching() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_updateMatching(t *testing.T) {
	type fields struct {
		mngClient *mongo.Client
		DbName    string
	}
	type args struct {
		ctx      context.Context
		matching Matching
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *mongo.UpdateResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repo{
				mngClient: tt.fields.mngClient,
				DbName:    tt.fields.DbName,
			}
			got, err := r.updateMatching(tt.args.ctx, tt.args.matching)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateMatching() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateMatching() got = %v, want %v", got, tt.want)
			}
		})
	}
}
