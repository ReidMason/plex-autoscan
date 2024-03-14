package api

import (
	"log/slog"
	"net/http"

	"github.com/ReidMason/plex-autoscan/internal/sonarr"
	"github.com/labstack/echo"
)

func (s Server) handleNotify(c echo.Context) error {
	serviceName := c.Param("service")

	s.log.Info("Received request", slog.String("ServiceName", serviceName))
	var body sonarr.SonarrWebhookBody
	err := c.Bind(&body)
	if err != nil {
		body := c.Request().Body
		s.log.Error("Failed to bind request body", slog.Any("error", err), slog.Any("body", body))
		return c.String(http.StatusBadRequest, "Invalid request body")
	}

	err = s.notificationHandler.ProcessNotification(body, serviceName)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "Request recieved from "+serviceName)
}
