package handlers

import (
	"bufio"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/gorilla/websocket"
)

type LogsHandler struct {
	upgrader websocket.Upgrader
}

func NewLogsHandler() *LogsHandler {
	return &LogsHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow connections from any origin (configure appropriately for production)
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// StreamLogs upgrades the connection to WebSocket and streams journalctl logs
func (h *LogsHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	nLines := 50 // default
	if n := r.URL.Query().Get("n"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			nLines = parsed
		}
	}

	since := r.URL.Query().Get("since") // optional timestamp

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Build journalctl command
	args := []string{
		"-u", "storage-api",
		"-f",           // follow (stream new logs)
		"-o", "json",   // JSON output format
		"--no-pager",   // don't use pager
		"-n", strconv.Itoa(nLines), // number of recent lines
	}

	if since != "" {
		args = append(args, "--since", since)
	}

	cmd := exec.Command("journalctl", args...)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to get stdout pipe: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "failed to start log stream"}`))
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start journalctl: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "failed to start journalctl"}`))
		return
	}

	// Channel to signal when to stop
	done := make(chan struct{})

	// Goroutine to detect client disconnect
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				close(done)
				return
			}
		}
	}()

	// Goroutine to kill process when done
	go func() {
		<-done
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Read and stream logs
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-done:
			return
		default:
			line := scanner.Text()
			if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
				log.Printf("WebSocket write error: %v", err)
				close(done)
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}

	// Wait for the command to finish
	cmd.Wait()
}
