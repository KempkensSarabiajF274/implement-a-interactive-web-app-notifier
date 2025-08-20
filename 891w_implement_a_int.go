package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Notification struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Time  string `json:"time"`
}

type NotificationRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var notifications = make([]Notification, 0)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/v1/notifications", handleNotifications)
	http.HandleFunc("/api/v1/notifications/", handleNotification)

	log.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(string(message))

		// Send notifications to client
		for _, notification := range notifications {
			jsonMessage, err := json.Marshal(notification)
			if err != nil {
				log.Println(err)
			} else {
				ws.WriteMessage(websocket.TextMessage, jsonMessage)
			}
		}
	}
}

func handleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var notificationRequest NotificationRequest
		err := json.NewDecoder(r.Body).Decode(&notificationRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		notification := Notification{
			ID:    fmt.Sprintf("%d", time.Now().UnixNano()),
			Title: notificationRequest.Title,
			Body:  notificationRequest.Body,
			Time:  time.Now().Format(time.RFC3339),
		}
		notifications = append(notifications, notification)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(notification)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func handleNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		keys := r.URL.Query()
		id := keys.Get("id")
		if id != "" {
			for _, notification := range notifications {
				if notification.ID == id {
					json.NewEncoder(w).Encode(notification)
					return
				}
			}
			http.Error(w, "Notification not found", http.StatusNotFound)
			return
		} else {
			json.NewEncoder(w).Encode(notifications)
		}
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}