package handler

import (
	"encoding/json"
	"net/http"

	"links-checker/internal/dto"
	"links-checker/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.healthCheck)
	mux.HandleFunc("/api/check", h.checkLinks)
	mux.HandleFunc("/api/report", h.generateReport)
}

// healthCheck просто проверяем, что сервер запущен и работает
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) checkLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not Allowed Method", http.StatusBadRequest)
		return
	}

	var req dto.CheckLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Links) == 0 {
		http.Error(w, "Links array is empty", http.StatusBadRequest)
		return
	}

	// пока мок тут бахнем ддля проверки
	resp := dto.CheckLinksResponse{
		Links: map[string]string{
			"google.com":       "available",
			"malformedlink.gg": "not available",
		},
		LinksNum: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) generateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.LinksList) == 0 {
		http.Error(w, "Links list is empty", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
