package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/store"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, dto.ErrorPayload(err.Error()))
}

func writeStoreError(w http.ResponseWriter, err error) {
	var missionConflict *store.MissionStartConflictError
	switch {
	case errors.As(err, &missionConflict):
		writeJSON(w, http.StatusConflict, dto.MissionConflictPayload(err.Error(), missionConflict.Conflicts))
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, store.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, store.ErrInvalidState):
		writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusBadRequest, err)
	}
}
