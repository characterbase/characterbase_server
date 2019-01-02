package api

import (
	"cbs/models"
	"cbs/services"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

// Service represents a resource service
type Service struct {
	Providers *Providers
	Config    *Config
}

// Services represents a group of resource services
type Services struct {
	Auth     services.Auth
	User     services.User
	Universe services.Universe
}

// Providers represents a collection of external connections
// (e.g. database, redis) necessary for the API to operate
type Providers struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// Middlewares represents a collection of API-specific middlewares
type Middlewares struct {
	UserSession  func(http.Handler) http.Handler
	Collaborator func(models.CollaboratorRole) func(http.Handler) http.Handler
}

// Config represents API settings loaded from a YAML configuration file
type Config struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	DatabaseURL   string `yaml:"database_url"`
	RedisURL      string `yaml:"redis_url"`
	MaxSessionAge string `yaml:"max_session_age"`
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
		},
	}
}
