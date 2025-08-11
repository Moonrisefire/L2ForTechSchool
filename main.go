package main

import (
	"Test/calendar"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var cal *calendar.Calendar

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cal = calendar.New()

	mux := http.NewServeMux()
	mux.Handle("/create_event", loggingMiddleware(http.HandlerFunc(createEventHandler)))
	mux.Handle("/update_event", loggingMiddleware(http.HandlerFunc(updateEventHandler)))
	mux.Handle("/delete_event", loggingMiddleware(http.HandlerFunc(deleteEventHandler)))
	mux.Handle("/events_for_day", loggingMiddleware(http.HandlerFunc(eventsForDayHandler)))
	mux.Handle("/events_for_week", loggingMiddleware(http.HandlerFunc(eventsForWeekHandler)))
	mux.Handle("/events_for_month", loggingMiddleware(http.HandlerFunc(eventsForMonthHandler)))

	log.Printf("Starting server on :%s", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func parseIntParam(r *http.Request, key string) (int, error) {
	val := r.FormValue(key)
	if val == "" {
		return 0, fmt.Errorf("missing parameter %s", key)
	}
	return strconv.Atoi(val)
}

func parseDateParam(r *http.Request, key string) (string, error) {
	val := r.FormValue(key)
	if val == "" {
		return "", fmt.Errorf("missing parameter %s", key)
	}
	_, err := time.Parse("2006-01-02", val)
	if err != nil {
		return "", fmt.Errorf("invalid date format %s", val)
	}
	return val, nil
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func createEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	userID, dateStr, eventText, err := parseCreateUpdateParams(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	id, err := cal.CreateEvent(userID, dateStr, eventText)
	if err != nil {
		handleBusinessError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"result": "event created", "id": id})
}

func updateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing parameter id"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	userID, dateStr, eventText, err := parseCreateUpdateParams(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err = cal.UpdateEvent(id, userID, dateStr, eventText)
	if err != nil {
		handleBusinessError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "event updated"})
}

func deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing parameter id"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	err = cal.DeleteEvent(id)
	if err != nil {
		handleBusinessError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "event deleted"})
}

func eventsForDayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	userID, err := parseIntParam(r, "user_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	dateStr, err := parseDateParam(r, "date")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	events, err := cal.EventsForDay(userID, dateStr)
	if err != nil {
		handleBusinessError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func eventsForWeekHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	userID, err := parseIntParam(r, "user_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	dateStr, err := parseDateParam(r, "date")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	events, err := cal.EventsForWeek(userID, dateStr)
	if err != nil {
		handleBusinessError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func eventsForMonthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	userID, err := parseIntParam(r, "user_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	dateStr, err := parseDateParam(r, "date")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	events, err := cal.EventsForMonth(userID, dateStr)
	if err != nil {
		handleBusinessError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func parseCreateUpdateParams(r *http.Request) (userID int, dateStr, eventText string, err error) {
	ct := r.Header.Get("Content-Type")
	if ct == "application/json" || (len(ct) >= 16 && ct[:16] == "application/json") {
		var body struct {
			UserID int    `json:"user_id"`
			Date   string `json:"date"`
			Event  string `json:"event"`
		}
		err = json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			return 0, "", "", fmt.Errorf("invalid json body")
		}
		userID = body.UserID
		dateStr = body.Date
		eventText = body.Event
	} else {
		err = r.ParseForm()
		if err != nil {
			return 0, "", "", fmt.Errorf("invalid form body")
		}
		userIDStr := r.FormValue("user_id")
		if userIDStr == "" {
			return 0, "", "", fmt.Errorf("missing user_id")
		}
		userID, err = strconv.Atoi(userIDStr)
		if err != nil {
			return 0, "", "", fmt.Errorf("invalid user_id")
		}
		dateStr = r.FormValue("date")
		if dateStr == "" {
			return 0, "", "", fmt.Errorf("missing date")
		}
		eventText = r.FormValue("event")
		if eventText == "" {
			return 0, "", "", fmt.Errorf("missing event")
		}
	}
	_, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid date format")
	}
	return userID, dateStr, eventText, nil
}

func handleBusinessError(w http.ResponseWriter, err error) {
	switch err {
	case calendar.ErrInvalidDate:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case calendar.ErrEventNotFound:
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.String(), time.Since(start))
	})
}
