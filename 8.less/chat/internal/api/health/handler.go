package health

import (
	"chat/pkg/logger"
	"net/http"
)

func CheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		logger.Log.Error("Error writing health check response", "error", err)
	}
}
