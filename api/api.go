package api

import (
	"cbs/models"
	"cbs/services"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
)

// Service represents a resource service
type Service struct {
	Providers *Providers
	Config    *Config
}

// Services represents a group of resource services
type Services struct {
	Auth      services.Auth
	User      services.User
	Universe  services.Universe
	Character services.Character
}

// Providers represents a collection of external connections
// (e.g. database, redis) necessary for the API to operate
type Providers struct {
	DB    *sqlx.DB
	Redis *redis.Client
}

// Middlewares represents a collection of API-specific middlewares
type Middlewares struct {
	UserSession  func(http.Handler) http.Handler
	Collaborator func(models.CollaboratorRole) func(http.Handler) http.Handler
	Universe     func(http.Handler) http.Handler
	Character    func(http.Handler) http.Handler
}

// Config represents API settings loaded from a YAML configuration file
type Config struct {
	Host               string   `yaml:"host"`
	Port               int      `yaml:"port"`
	DatabaseURL        string   `yaml:"database_url"`
	RedisURL           string   `yaml:"redis_url"`
	MaxSessionAge      string   `yaml:"max_session_age"`
	CharacterPageLimit int      `yaml:"character_page_limit"`
	AllowedOrigins     []string `yaml:"allowed_origins"`
}

// Server represents an API server with a loaded configuration and set of providers
type Server struct {
	*chi.Mux
	Providers   *Providers
	Config      *Config
	Services    *Services
	Middlewares *Middlewares
}

// NewServer creates a new API server
func NewServer(config Config, providers *Providers, services *Services) *Server {
	return &Server{
		Mux:       chi.NewMux(),
		Config:    &config,
		Providers: providers,
		Services:  services,
		Middlewares: &Middlewares{
			UserSession:  MwUserSession(services),
			Collaborator: MwCollaborator(services),
			Universe:     MwUniverse(services),
			Character:    MwCharacter(services),
		},
	}
}
