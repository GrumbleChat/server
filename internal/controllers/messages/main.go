package messagesController

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/ksuid"
	"gitlab.com/grumblechat/server/pkg/channel"
	bolt "go.etcd.io/bbolt"
)

func BindRoutes(db *bolt.DB, router *echo.Group) {
	// router.GET("/", listHandler(db))
	router.POST("/", createHandler(db))
}

func validateChannel(db *bolt.DB, id ksuid.KSUID) *echo.HTTPError {
	c, err := channel.Find(db, id)

	if (err != nil) {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if (c == nil) {
		return echo.NewHTTPError(http.StatusNotFound, "Channel ID not recognized")
	}

	if (c.GetType() != "text") {
		return echo.NewHTTPError(http.StatusBadRequest, "Messages are only available for 'text' channel types")
	}

	return nil
}