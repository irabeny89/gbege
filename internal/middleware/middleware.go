package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/irabeny89/gbege/internal/api"
	"github.com/irabeny89/gbege/internal/logger"
)

const REQUEST_ID_KEY = "request_id"

func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid, err := uuid.NewV7()
		if err != nil {
			logger.Log.Error("Failed to generate request ID", "err", err)
			api.Fail(w, http.StatusInternalServerError, "Internal server error", err)
			return
		}
		requestID := uuid.String()

		// Add to context so your DB logic can log it
		ctx := context.WithValue(r.Context(), REQUEST_ID_KEY, requestID)

		// Set in response header so the client/frontend can see it
		w.Header().Set("X-Request-ID", requestID)

		logger.Log.Info("Request ID", "req_id", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
