package http

import (
	"bytes"
	"fmt"
	"github.com/shamank/edutour-backend/apigateway/pkg/logger/sl"
	"github.com/valyala/fasthttp"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

type Services struct {
	AuthServiceAddr string
	DataServiceAddr string
}

type Handler struct {
	logger   *slog.Logger
	services Services
}

func NewHandler(logger *slog.Logger, services Services) *Handler {
	return &Handler{
		logger:   logger,
		services: services,
	}
}

func (h *Handler) InitHandle() func(ctx *fasthttp.RequestCtx) {
	router := func(ctx *fasthttp.RequestCtx) {
		switch path := string(ctx.Path()); true {

		case strings.HasPrefix(path, "/api/v1/auth") ||
			strings.HasPrefix(path, "/api/v1/users") ||
			strings.HasPrefix(path, "/swagger"):

			h.proxyRequest(ctx, h.services.AuthServiceAddr)

		case strings.HasPrefix(path, "/api/v1") ||
			strings.HasPrefix(path, "/docs") ||
			strings.HasPrefix(path, "/openapi.json"):

			userData, err := h.GetUserInfo(ctx)

			if err != nil {
				ctx.QueryArgs().Del("user_id")
				ctx.QueryArgs().Del("user_role")

				ctx.QueryArgs().SetUint("user_role", 0)
			} else {
				ctx.QueryArgs().SetUint("user_id", userData.ID)
				ctx.QueryArgs().SetUint("user_role", userData.Role)
				ctx.QueryArgs().SetUint("user_id_to_get", userData.ID)
			}

			h.proxyRequest(ctx, h.services.DataServiceAddr)

		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}

	return router
}

func (h *Handler) proxyRequest(ctx *fasthttp.RequestCtx, to string) {
	const op = "Delivery.Http.Handler"
	logger := h.logger.With(slog.String("op", op), slog.String("ctx", ctx.String()))

	logger.Info("get request to: " + to + string(ctx.Path()))

	client := &http.Client{}

	// Чтение Request Body
	reqBody := bytes.NewReader(ctx.Request.Body())

	link, err := url.Parse(to)
	if err != nil {
		logger.Error("failed on parse url: ", err.Error())
		return
	}
	link.Path = string(ctx.Path())

	req, err := http.NewRequest(string(ctx.Request.Header.Method()), link.String(), reqBody)
	if err != nil {
		logger.Error("failed to create NewRequest", sl.Err(err))
		return
	}

	q := req.URL.Query()
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		fmt.Println("key: " + string(key) + " value: " + string(value))
		q.Set(string(key), string(value))
	})
	req.URL.RawQuery = q.Encode()

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Отправка HTTP-запроса и получение ответа
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("error occurred while sending the HTTP request", sl.Err(err))
		return
	}

	defer resp.Body.Close()

	// Чтение тела ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("error occurred while reading the HTTP response", sl.Err(err))
		return
	}

	// Вывод тела ответа
	ctx.SetStatusCode(resp.StatusCode)
	ctx.SetBody(body)

	for key := range resp.Header {
		ctx.Response.Header.Set(key, resp.Header.Get(key))
	}

}
