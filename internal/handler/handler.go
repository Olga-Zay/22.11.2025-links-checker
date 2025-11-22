package handler

import (
	"encoding/json"
	"log"
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
		//в нормальной задаче тут был бы логгер
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Links) == 0 {
		log.Printf("Links array is empty in req: %v", req)
		http.Error(w, "Links array is empty", http.StatusBadRequest)
		return
	}

	if len(req.Links) > 100 {
		log.Printf("Too many links in task in req: %v", req)
		http.Error(w, "Too many links in task", http.StatusBadRequest)
		return
	}

	taskID, err := h.service.CheckLinks(req.Links)
	if err != nil {
		log.Printf("Failed to create check task: %v", err)
		http.Error(w, "Failed to create check task", http.StatusInternalServerError)
		return
	}

	checkTask, err := h.service.GetLinkCheckTaskResults(taskID)
	if err != nil {
		log.Printf("Failed to get check task: %v", err)
		http.Error(w, "Failed to get check task", http.StatusInternalServerError)
		return
	}

	linksMap := make(map[string]string)
	for _, link := range checkTask.Links {
		linksMap[link.URL] = string(link.Status)
	}

	resp := dto.CheckLinksResponse{
		Links:    linksMap,
		LinksNum: taskID,
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
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.LinksList) == 0 {
		log.Printf("Links list is empty in req: %v", req)
		http.Error(w, "Links list is empty", http.StatusBadRequest)
		return
	}

	if len(req.LinksList) > 100 {
		log.Printf("Links list is too big in req: %v", req)
		http.Error(w, "Links list is too big", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
