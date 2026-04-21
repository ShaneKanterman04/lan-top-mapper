package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"lan-topology-mapper/internal/discovery"
	"lan-topology-mapper/internal/models"
	"lan-topology-mapper/internal/store"
)

type Server struct {
	store   *store.Store
	scanner *discovery.Scanner
	static  string

	mu         sync.Mutex
	activeScan bool
}

func NewServer(store *store.Store, scanner *discovery.Scanner, staticDir string) *Server {
	return &Server{store: store, scanner: scanner, static: staticDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/scans", s.handleScans)
	mux.HandleFunc("/api/scans/", s.handleScanByID)
	mux.HandleFunc("/api/devices", s.handleDevices)
	mux.HandleFunc("/api/devices/", s.handleDeviceByID)
	mux.HandleFunc("/api/summary", s.handleSummary)
	mux.HandleFunc("/api/topology", s.handleTopology)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	if hasStaticBuild(s.static) {
		mux.Handle("/", spaHandler(s.static))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{
				"message": "frontend build not found; run the web build or use Docker image build",
			})
		})
	}
	return withCORS(mux)
}

func (s *Server) handleScans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	s.mu.Lock()
	if s.activeScan {
		s.mu.Unlock()
		writeJSON(w, http.StatusConflict, map[string]string{"error": "scan already running"})
		return
	}
	s.activeScan = true
	s.mu.Unlock()

	plan, err := s.scanner.PreparePlan()
	if err != nil {
		s.finishScanFlag()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	scanID, err := s.store.CreateScanRun(r.Context(), plan.Subnet, plan.GatewayIP, models.ScanDiagnostics{
		Mode:             plan.Mode,
		AvailableSources: s.scanner.Capabilities().AvailableSources,
		TotalTargets:     len(plan.Targets),
	})
	if err != nil {
		s.finishScanFlag()
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	go s.runScan(scanID, plan)
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "started", "scanId": scanID, "subnet": plan.Subnet, "mode": plan.Mode})
}

func (s *Server) runScan(scanID int64, plan discovery.Plan) {
	defer s.finishScanFlag()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := s.scanner.Scan(ctx, plan)
	if err != nil {
		_ = s.store.FailScanRun(context.Background(), scanID, err)
		return
	}
	if err := s.store.ReplaceScanObservations(context.Background(), scanID, result.Observations); err != nil {
		_ = s.store.FailScanRun(context.Background(), scanID, err)
		return
	}
	if err := s.store.CompleteScanRun(context.Background(), scanID, result.Diagnostics); err != nil {
		log.Printf("complete scan run %d: %v", scanID, err)
	}
}

func (s *Server) finishScanFlag() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeScan = false
}

func (s *Server) handleScanByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id, err := parseID(r.URL.Path, "/api/scans/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	scan, err := s.store.GetScanRun(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if scan == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "scan not found"})
		return
	}
	writeJSON(w, http.StatusOK, scan)
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	devices, err := s.store.ListDevices(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, devices)
}

func (s *Server) handleDeviceByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id, err := parseID(r.URL.Path, "/api/devices/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	device, err := s.store.GetDeviceDetail(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if device == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	summary, err := s.store.Summary(r.Context(), s.scanner.Capabilities())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleTopology(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	topology, err := s.store.Topology(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, topology)
}

func parseID(path, prefix string) (int64, error) {
	value := strings.TrimPrefix(path, prefix)
	value = strings.Trim(value, "/")
	if value == "" {
		return 0, errors.New("missing id")
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func spaHandler(staticDir string) http.Handler {
	fs := http.FileServer(http.Dir(staticDir))
	indexPath := filepath.Join(staticDir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		relPath := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		candidate := filepath.Join(staticDir, relPath)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}

func hasStaticBuild(staticDir string) bool {
	info, err := os.Stat(filepath.Join(staticDir, "index.html"))
	return err == nil && !info.IsDir()
}
