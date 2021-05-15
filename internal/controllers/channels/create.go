package channelsController

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"gitlab.com/grumblechat/server/pkg/channel"
	bolt "go.etcd.io/bbolt"
)

func getTypeFromBody(ctx echo.Context) (string, error) {
	type typeEnvelope struct {
		Type string `json:"type" validate:"oneof=text voice,required"`
	}

	// get raw bytes from body
	var (bodyBytes []byte; err error)
	if bodyBytes, err = ioutil.ReadAll(ctx.Request().Body); err != nil {
		return "", echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// add body back into request
	ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// unmarshal JSON from body
	envelope := new(typeEnvelope)
	if err = json.Unmarshal(bodyBytes, envelope); err != nil {
		return "", echo.NewHTTPError(http.StatusInternalServerError, err.Error()) 
	}

	// validate type
	if err = ctx.Validate(envelope); err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// return type and no error
	return envelope.Type, nil
}

func createHandler(db *bolt.DB) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var newChannel channel.Channel

		// unmarshal raw body and write back to request
		channelType, err := getTypeFromBody(ctx)
		if err != nil { return err }

		// voice channel
		if channelType == "voice" {
			newChannel = channel.NewVoice()
			if err := ctx.Bind(newChannel); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
		}

		// text channel
		if channelType == "text" {
			newChannel = channel.NewText()
			if err := ctx.Bind(newChannel); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
		}

		// validate channel
		if err = ctx.Validate(newChannel); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// persist channel
		if err := newChannel.Save(db); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusCreated, newChannel)
	}
}