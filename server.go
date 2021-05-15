package main

import (
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
	sentryEcho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	bolt "go.etcd.io/bbolt"

	"gitlab.com/grumblechat/server/internal/config"
	"gitlab.com/grumblechat/server/internal/controllers/channels"
	messagesController "gitlab.com/grumblechat/server/internal/controllers/messages"
	"gitlab.com/grumblechat/server/internal/validation"
)

func initDB(path string) *bolt.DB {
	// open BoltDB
	dbPath := fmt.Sprintf("%s/grumble.db", path)
	db, err := bolt.Open(dbPath, 0666, nil)

	if err != nil {
		panic("Failed to open database")
	}

	// ensure that buckets exist
	err = db.Update(func(tx *bolt.Tx) error {
		// channels
		_, err := tx.CreateBucketIfNotExists([]byte("channels"))
		if (err != nil) { return err }

		return nil
	})

	if (err != nil) {
		panic("Failed to migrate DB")
	}

	return db
}

func main() {
	// load config
	config := config.Load()

	// initialize Sentry client
	err := sentry.Init(sentry.ClientOptions{
		Dsn: config.SentryDSN,
	})
	if err != nil {
		log.Fatalf("Sentry initialization failed: %v\n", err)
	}

	// init framework
	app := echo.New()
	app.Validator = validation.Echo()
	app.Pre(middleware.AddTrailingSlash())

	// register global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	// report errors to sentry
	if config.EnableSentry {
		app.Use(sentryEcho.New(sentryEcho.Options{
			Repanic: true,
		}))
	}

	// load database
	db := initDB(config.Storage.Database)
	defer db.Close()

	// bind controller routes
	channelsController.BindRoutes(db, app.Group("/channels"))
	messagesController.BindRoutes(db, app.Group("/channels/:channelID/messages"))

	// start server
	app.Start(fmt.Sprintf("%s:%d", config.Host, config.Port))
}
