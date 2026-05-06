// mockexchange is a minimal ActiveSync endpoint that mimics Exchange CAS responses.
// It is used for demo deployments where a real Exchange server is not available.
// It accepts any credentials, validates the EAS wire format, and returns realistic
// but empty responses so the full SyncGuard auth flow can be demonstrated end-to-end.
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/Microsoft-Server-ActiveSync", handleEAS)
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/", handleRoot)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  90 * time.Second,
	}

	slog.Info("mock Exchange listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

// handleEAS handles all /Microsoft-Server-ActiveSync requests.
// It mirrors the response headers Exchange uses and returns empty but valid
// responses for the most common EAS commands.
func handleEAS(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("Cmd")
	deviceID := r.URL.Query().Get("DeviceId")
	userAgent := r.UserAgent()

	slog.Info("EAS request",
		"method", r.Method,
		"cmd", cmd,
		"device_id", deviceID,
		"user_agent", userAgent,
		"remote_addr", r.RemoteAddr,
	)

	// Standard Exchange CAS response headers
	w.Header().Set("MS-Server-ActiveSync", "16.1")
	w.Header().Set("MS-ASProtocolVersions", "2.5,12.0,12.1,14.0,14.1,16.0,16.1")
	w.Header().Set("MS-ASProtocolCommands",
		"Sync,SendMail,SmartForward,SmartReply,GetAttachment,GetHierarchy,CreateCollection,DeleteCollection,MoveCollection,FolderSync,FolderCreate,FolderDelete,FolderUpdate,MoveItems,GetItemEstimate,MeetingResponse,Search,Settings,Ping,ItemOperations,Provision,ResolveRecipients,ValidateCert")
	w.Header().Set("Cache-Control", "private")

	switch r.Method {
	case http.MethodOptions:
		// OPTIONS — client probing protocol versions
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		handleEASCommand(w, r, cmd)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleEASCommand(w http.ResponseWriter, r *http.Request, cmd string) {
	switch cmd {
	case "Ping":
		// Ping — heartbeat, keep connection alive
		// Status 1 = no changes, client should re-ping
		w.Header().Set("Content-Type", "application/vnd.ms-sync.wbxml")
		w.WriteHeader(http.StatusOK)
		// Minimal valid Ping response WBXML: Status=1 (NoChanges)
		w.Write([]byte{0x03, 0x01, 0x6A, 0x00, 0x45, 0x4C, 0x03, 0x31, 0x00, 0x01, 0x01})

	case "Provision":
		// Provision — device policy negotiation
		// Status 1 = Success, no policy enforcement for demo
		w.Header().Set("Content-Type", "application/vnd.ms-sync.wbxml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{0x03, 0x01, 0x6A, 0x00})

	case "FolderSync":
		// FolderSync — client requesting folder list
		w.Header().Set("Content-Type", "application/vnd.ms-sync.wbxml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{0x03, 0x01, 0x6A, 0x00})

	case "Sync":
		// Sync — mail/calendar/contacts sync
		// Return 200 with empty sync response (no new items)
		w.Header().Set("Content-Type", "application/vnd.ms-sync.wbxml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{0x03, 0x01, 0x6A, 0x00})

	case "GetHierarchy":
		w.Header().Set("Content-Type", "application/vnd.ms-sync.wbxml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{0x03, 0x01, 0x6A, 0x00})

	case "Settings":
		w.Header().Set("Content-Type", "application/vnd.ms-sync.wbxml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{0x03, 0x01, 0x6A, 0x00})

	case "":
		// No Cmd param — initial OPTIONS-style POST
		w.WriteHeader(http.StatusOK)

	default:
		// Unknown command — return 200 with empty body
		// Real Exchange would return 400 but for demo purposes accept everything
		slog.Warn("unknown EAS command", "cmd", cmd)
		w.Header().Set("Content-Type", "application/vnd.ms-sync.wbxml")
		w.WriteHeader(http.StatusOK)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "mock-exchange",
		"version": "Exchange Server 2019 (mock)",
	})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "SyncGuard Mock Exchange — demo endpoint")
}
