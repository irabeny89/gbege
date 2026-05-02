package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/irabeny89/gbege/internal/api"
	"github.com/irabeny89/gbege/internal/logger"
)

func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid, err := uuid.NewV7()
		if err != nil {
			logger.Log.Error("Failed to generate request ID", "err", err)
			api.Fail(w, http.StatusInternalServerError, "Internal server error", err)
			return
		}
		requestID := uuid.String()

		logger.Log.Info(
			"Request received",
			"ip_addr", r.RemoteAddr,
			"path", r.URL.Path,
			"method", r.Method,
			"req_id", requestID,
		)
		logger.WithAttrs("req_id", requestID)

		// Set in response header so the client/frontend can see it
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}
