package refresh

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	resp "medods/internal/lib/api/response"
	"medods/internal/lib/jwt"
	"medods/internal/lib/logger/sl"
	"medods/internal/storage/models"
	"net/http"
	"time"
)

type Request struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Response struct {
	resp.Response
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokensRefresher interface {
	VerifyRefreshTokenHash(token string, timeout time.Duration) (*models.Authorization, error)
	SaveRefreshTokenHash(authToken *models.Authorization, timeout time.Duration) error
}

func New(log *slog.Logger, tokenRefresher TokensRefresher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.refresh.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", chi.URLParam(r, "request_id")),
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

		authInfo, err := tokenRefresher.VerifyRefreshTokenHash(req.RefreshToken, 5*time.Second)
		if err != nil {
			log.Error("invalid refresh token", sl.Err(err))

			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, resp.Error("invalid refresh token"))

			return
		}

		newAccessToken, err := jwt.NewAccessToken(authInfo.UserGUID.String())
		if err != nil {
			log.Error("failed to generate new access token", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to generate new access token"))

			return
		}

		newRefreshToken, newRefreshTokenHash, err := jwt.NewRefreshToken()
		if err != nil {
			log.Error("failed to generate new refresh token", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to generate new refresh token"))

			return
		}

		authInfo.RefreshTokenHash = newRefreshTokenHash
		err = tokenRefresher.SaveRefreshTokenHash(authInfo, 5*time.Second)
		if err != nil {
			log.Error("failed to save new refresh token hash", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to save new refresh token"))
			return
		}

		responseOK(w, r, newAccessToken, newRefreshToken)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	render.JSON(w, r, Response{
		Response:     resp.OK(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
