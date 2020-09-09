package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoDbConfig Config
	client        *mongo.Client
)

type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	DriverName      string
	DbHost          string
	DbPort          string
	DbName          string
}

// Addr returns server address in the form of Host:Port localhost:8080.
func (sc Config) Addr() string {
	return sc.Host + ":" + strconv.Itoa(sc.Port)
}

func main() {
	config := newConfig()
	if err := startAPIServerAndWait(config); err != nil {
		log.Fatal(err)
	}
}

func newConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            8090,
		DriverName:      "mongodb",
		DbHost:          "localhost", //# aws winawin mongodb public ip
		DbPort:          "27017",
		DbName:          "winawin_test",
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
		ShutdownTimeout: time.Second * 5,
	}
}

func NewMongoClient(conf Config) (*mongo.Client, error) {
	mongoDbConfig = conf

	mongoURI := fmt.Sprintf("%s://%s:%s/?retryWrites=false",
		mongoDbConfig.DriverName, mongoDbConfig.DbHost, mongoDbConfig.DbPort)
	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	if client, err = mongo.Connect(context.TODO(), clientOptions); err != nil {
		return client, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return client, err
	}

	log.Printf("Connected to MongoDB!")
	log.Printf("    Connection URL: %s\n\n", mongoURI)
	return client, nil
}

func startAPIServerAndWait(config Config) error {
	mongoClient, err := NewMongoClient(config)
	if err != nil {
		return err
	}

	defer func() {
		log.Printf("main : Database Stopping : %s", config.DbHost)
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()


	log.Printf("int-matching : API server listening on %s", config.Addr())

	_, err = startAPIServer(config, mongoClient)
	if err != nil {
		return err
	}
	return nil
}

func startAPIServer(cfg Config,	mongoClient *mongo.Client) (*http.Server, error) {

	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	matchingServer := NewServer("development", cfg.DbName, mongoClient)
	r.Mount("/api/v1/matching", matchingServer.Router)

	server := http.Server{
		Addr:         cfg.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		return nil, err
	}
	return &server, nil
}
