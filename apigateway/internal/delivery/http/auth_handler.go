package http

import (
	"encoding/json"
	"errors"
	"github.com/shamank/edutour-backend/apigateway/pkg/logger/sl"
	"github.com/valyala/fasthttp"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type UserData struct {
	ID   int `json:"id"`
	Role int `json:"role"`
}

func (h *Handler) GetUserInfo(ctx *fasthttp.RequestCtx) (UserData, error) {
	const op = "Delivery.Http.AuthHandler"
	logger := h.logger.With(slog.String("op", op), slog.String("ctx", ctx.String()))

	link, err := url.Parse(h.services.AuthServiceAddr)
	if err != nil {
		logger.Error("failed on parse url", sl.Err(err))
		return UserData{}, err
	}

	link.Path = "/api/v1/auth/verify"

	req, err := http.NewRequest(http.MethodGet, link.String(), nil)
	if err != nil {
		logger.Error("failed to create NewRequest", sl.Err(err))
		return UserData{}, err
	}
	client := &http.Client{}

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("error occurred while sending the HTTP request", sl.Err(err))
		return UserData{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("error occurred while reading the HTTP response: ", sl.Err(err))
		return UserData{}, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Warn("user is not authorized")

		return UserData{}, errors.New("user is not authorized")
	}

	var userData UserData

	err = json.Unmarshal(body, &userData)
	if err != nil {
		logger.Error("error occurred while unmarshall json", sl.Err(err))
		return UserData{}, err
	}

	return userData, nil
}
