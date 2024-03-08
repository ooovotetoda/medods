package save

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"log/slog"
	"medods/internal/lib/jwt"
	"medods/internal/lib/logger/sl"
	"medods/internal/storage/models"
	"net/http"
	"time"

	resp "medods/internal/lib/api/response"
)

type Request struct {
	GUID string `json:"guid" validate:"required"`
}

type Response struct {
	resp.Response
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenSaver interface {
	SaveRefreshTokenHash(authToken *models.Authorization, timeout time.Duration) error
}

func New(log *slog.Logger, refreshTokenSaver RefreshTokenSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("failed to validate request", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		parsedGUID, err := uuid.Parse(req.GUID)
		if err != nil {
			log.Error("failed to parse GUID", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid UUID format"))

			return
		}

		accessToken, err := jwt.NewAccessToken(req.GUID)
		if err != nil {
			log.Error("failed to generate access token", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		refreshToken, refreshTokenHash, err := jwt.NewRefreshToken()
		if err != nil {
			log.Error("failed to save refresh token hash")

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		authInfo := &models.Authorization{
			UserGUID:         parsedGUID,
			RefreshTokenHash: refreshTokenHash,
		}

		err = refreshTokenSaver.SaveRefreshTokenHash(authInfo, 5*time.Second)

		responseOK(w, r, accessToken, refreshToken)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	render.JSON(w, r, Response{
		Response:     resp.OK(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
