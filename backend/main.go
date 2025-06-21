package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	id     string
	events chan Event
	done   chan bool
}

type Event struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
}

var hub = &Hub{
	clients:    make(map[string]*Client),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	broadcast:  make(chan Event),
}

func (h *Hub) run() {
	// Send periodic heartbeat messages
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.broadcast <- Event{
					Type:      "heartbeat",
					Message:   "Server heartbeat",
					Timestamp: time.Now().Format(time.RFC3339),
				}
			}
		}
	}()

	for {
		select {
		case client := <-h.register:
			h.clients[client.id] = client
			log.Printf("Client %s connected. Total clients: %d", client.id, len(h.clients))

			welcome := Event{
				Type:      "welcome",
				Message:   "Connected to SSE server",
				Timestamp: time.Now().Format(time.RFC3339),
			}
			select {
			case client.events <- welcome:
			case <-client.done:
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.events)
				log.Printf("Client %s disconnected. Total clients: %d", client.id, len(h.clients))
			}

		case event := <-h.broadcast:
			for id, client := range h.clients {
				select {
				case client.events <- event:
				case <-client.done:
					delete(h.clients, id)
					close(client.events)
				}
			}
		}
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	client := &Client{
		id:     clientID,
		events: make(chan Event, 10),
		done:   make(chan bool),
	}

	hub.register <- client

	defer func() {
		hub.unregister <- client
		client.done <- true
	}()

	ctx := r.Context()
	go func() {
		<-ctx.Done()
		client.done <- true
	}()

	for {
		select {
		case event := <-client.events:
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("Error marshaling event: %v", err)
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()

		case <-client.done:
			return
		}
	}
}

func triggerHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	event := Event{
		Type:      "ping",
		Message:   "Ping received from client",
		Data:      req,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	hub.broadcast <- event

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"time":    time.Now().Format(time.RFC3339),
		"clients": len(hub.clients),
	})
}

func main() {
	go hub.run()

	http.HandleFunc("/events", sseHandler)
	http.HandleFunc("/trigger", triggerHandler)
	http.HandleFunc("/health", healthHandler)

	port := "8080"
	log.Printf("Server starting on port %s", port)
	log.Printf("SSE endpoint: http://localhost:%s/events", port)
	log.Printf("Trigger endpoint: http://localhost:%s/trigger", port)
	log.Printf("Health endpoint: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
