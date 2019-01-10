package main

import (
	"cbs/api"
	"cbs/api/auth"
	"cbs/api/characters"
	"cbs/api/universes"
	"cbs/api/users"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/go-redis/redis"

	"github.com/go-chi/chi/middleware"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	yaml "gopkg.in/yaml.v2"
)

var (
	configPath = flag.String("c", "config.yaml", "Path to the configuration file")
)

// loadConfig loads configuration from a YAML configuration file
func loadConfig(path string) (*api.Config, error) {
	var config *api.Config
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("failed to read configuration file")
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.New("failed to parse configuration file")
	}
	return config, nil
}

func newServices(providers *api.Providers, config *api.Config) *api.Services {
	return &api.Services{
		Auth:      &auth.Service{Providers: providers, Config: config},
		User:      &users.Service{Providers: providers, Config: config},
		Universe:  &universes.Service{Providers: providers, Config: config},
		Character: &characters.Service{Providers: providers, Config: config},
	}
}

func newServer(config api.Config, providers *api.Providers) *api.Server {
	services := newServices(providers, &config)
	server := api.NewServer(config, providers, services)

	// Mount the API middleware
	server.Use(middleware.Logger)

	// Mount the API routers
	server.Mount("/", auth.NewRouter(server))
	server.Mount("/users", users.NewRouter(server))
	server.Mount("/universes", universes.NewRouter(server))
	server.Mount("/universes/{universeID}/characters", characters.NewRouter(server))

	return server
}

func main() {
	// Parse the command-line flags
	flag.Parse()

	// Load API configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		panic(err)
	}

	// Connect to the database
	log.Printf("Connecting to database... (url: %v)\n", config.DatabaseURL)
	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		panic(err)
	}
	log.Printf("Database connection OK\n")

	// Connect to the Redis store
	log.Printf("Connecting to Redis... (url: %v)\n", config.RedisURL)
	redisdb := redis.NewClient(&redis.Options{Addr: config.RedisURL})
	_, err = redisdb.Ping().Result()
	if err != nil {
		panic(err)
	}
	log.Printf("Redis connection OK\n")

	// Setup the API providers
	providers := &api.Providers{DB: db, Redis: redisdb}

	// Create the API server
	server := newServer(*config, providers)

	// Start the API server
	address := fmt.Sprintf("%v:%d", config.Host, config.Port)
	log.Printf("CharacterBase API is now listening on %v...\n", address)
	if err := http.ListenAndServe(address, server); err != nil {
		log.Fatalln(err)
	}
}
