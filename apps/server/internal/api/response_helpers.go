package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"robot-center/apps/server/internal/store"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{
		"error": err.Error(),
	})
}

func writeStoreError(w http.ResponseWriter, err error) {
	var missionConflict *store.MissionStartConflictError
	switch {
	case errors.As(err, &missionConflict):
		conflicts := make([]map[string]string, 0, len(missionConflict.Conflicts))
		for _, conflict := range missionConflict.Conflicts {
			conflicts = append(conflicts, map[string]string{
				"robotCode":         conflict.RobotCode,
				"activeMissionCode": conflict.ActiveMissionCode,
			})
		}
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":     err.Error(),
			"conflicts": conflicts,
		})
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

func withRequestHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robot-Center", "app-server")
		next.ServeHTTP(w, r)
	})
}
