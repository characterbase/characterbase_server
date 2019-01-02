package integration

import (
	"cbs/api"
	"cbs/api/auth"
	"cbs/api/users"
	"cbs/dtos"
	"cbs/models"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/buger/jsonparser"

	"github.com/go-redis/redis"

	"github.com/jinzhu/gorm"
	yaml "gopkg.in/yaml.v2"
)

var (
	configPath = flag.String("tc", "config.yaml", "Path to the testing configuration file")

	server *api.Server

	// Test references
	userA         *models.User
	userB         *models.User
	userAPassword = "elliptical"
	userBPassword = "twinsisters"
	userASessKey  string
)

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

func apiRequest(method, route string, body io.Reader) (*http.Request, *httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(method, route, body)
	if err != nil {
		return nil, nil, err
	}
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	return req, rr, nil
}

func compareUser(a *models.User, b *models.User) bool {
	if a.DisplayName != b.DisplayName {
		return false
	}
	if a.Email != b.Email {
		return false
	}
	if a.ID != b.ID {
		return false
	}
	return true
}

func loginUser(r *http.Request, user *models.User) (*http.Cookie, error) {
	w := httptest.NewRecorder()
	if err := server.Services.Auth.Login(user, w); err != nil {
		return nil, err
	}
	sesscookie := findCookie("user_session", w.Result().Cookies())
	if sesscookie == nil {
		return nil, errors.New("login error: session not found")
	}
	r.AddCookie(sesscookie)
	return sesscookie, nil
}

func testAPIResponse(t *testing.T, res *httptest.ResponseRecorder, want interface{}, status int, normid bool) {
	if res.Code != status {
		t.Errorf("got response status %v; want %v", res.Code, status)
	}
	if fmt.Sprintf("%T", want) == "api.ErrorCode" {
		var reserr api.Error
		if err := json.Unmarshal(res.Body.Bytes(), &reserr); err != nil {
			t.Fatal("failed to unmarshal response error")
		}
		if reserr.Code != (want.(api.ErrorCode)) {
			t.Errorf("got error code %v; want %v", reserr.Code, (want.(api.ErrorCode)))
		}
	} else {
		var wantres string
		if fmt.Sprintf("%T", want) == "string" {
			wantres = want.(string)
		} else {
			parsed, err := json.Marshal(want)
			if err != nil {
				t.Fatal("failed to marshal wanted response")
			}
			wantres = string(parsed)
		}

		// A slightly hacky method to zero ID from responses for easier testing
		normwant := res.Body.Bytes()
		if normid {
			norm, err := jsonparser.Set(res.Body.Bytes(), []byte("\"\""), "id")
			if err != nil {
				t.Fatalf("failed to normalize response (remove id)")
			}
			normwant = norm
		}
		if string(normwant) != wantres {
			t.Errorf("got response %v; want %v", string(normwant), wantres)
		}
	}
}

func migrateDatabase(db *gorm.DB) {
	db.AutoMigrate(&models.User{})
}

func generateTestData() error {
	// Generate test users
	userA = server.Services.User.New(dtos.ReqCreateUser{
		DisplayName: "john",
		Email:       "john@gmail.com",
		Password:    userAPassword})
	if err := server.Services.User.Save(userA); err != nil {
		return err
	}
	userB = server.Services.User.New(dtos.ReqCreateUser{
		DisplayName: "mark",
		Email:       "mark@yahoo.com",
		Password:    userBPassword})
	if err := server.Services.User.Save(userB); err != nil {
		return err
	}

	return nil
}

func setup() error {
	// Parse the flags
	flag.Parse()

	// Read the configuration file
	config, err := loadConfig(*configPath)
	if err != nil {
		return err
	}

	// Start the database
	db, err := gorm.Open("postgres", config.DatabaseURL)
	if err != nil {
		return err
	}
	migrateDatabase(db)

	// Connect to the Redis store
	redisdb := redis.NewClient(&redis.Options{Addr: config.RedisURL, DB: 1})
	if _, err := redisdb.Ping().Result(); err != nil {
		return err
	}

	// Create the API server
	providers := &api.Providers{DB: db, Redis: redisdb}
	services := &api.Services{
		Auth: &auth.Service{Config: config, Providers: providers},
		User: &users.Service{Config: config, Providers: providers},
	}
	server = api.NewServer(*config, providers, services)

	// Mount the server routes
	server.Mount("/users", users.NewRouter(server))
	server.Mount("/", auth.NewRouter(server))

	// Generate the test data
	if err := generateTestData(); err != nil {
		return err
	}

	return nil
}

func teardown() error {
	if err := server.Providers.DB.Delete(&models.User{}).Error; err != nil {
		return err
	}

	// Clean the Redis database
	if err := server.Providers.Redis.FlushAll().Err(); err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		panic(err)
	}
	code := m.Run()
	if err := teardown(); err != nil {
		panic(err)
	}
	os.Exit(code)
}
