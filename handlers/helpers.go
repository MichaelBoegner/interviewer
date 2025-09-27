package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func ValidateInterviewStatusTransition(currentStatus, nextStatus string) error {
	validTransitions := map[string][]string{
		"not_started": {"active"},
		"active":      {"paused", "finished"},
		"paused":      {"active"},
		"finished":    {},
	}

	allowed, ok := validTransitions[currentStatus]
	if !ok {
		return fmt.Errorf("invalid current state: %s", currentStatus)
	}

	for _, a := range allowed {
		if a == nextStatus {
			return nil
		}
	}
	return fmt.Errorf("invalid state transition: %s → %s", currentStatus, nextStatus)
}

func GetPathID(r *http.Request, prefix string, logger *slog.Logger) (int, error) {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.Trim(path, "/")

	if path == "" {
		logger.Error("getPathID returned empty string")
		err := errors.New("missing or invalid url param")
		return 0, err
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		logger.Error("getPathID failed", "error", err)
		return 0, err
	}

	return id, nil
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	if code != 0 {
		w.WriteHeader(code)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	if code != 0 {
		w.WriteHeader(code)
	}

	respBody := ReturnVals{
		Error: msg,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Write(data)
}
