package calendar

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrInvalidDate   = errors.New("invalid date format")
)

type Event struct {
	ID     int       `json:"id"`
	UserID int       `json:"user_id"`
	Date   time.Time `json:"date"`
	Text   string    `json:"event"`
}

type Calendar struct {
	mu     sync.RWMutex
	events map[int]*Event
	nextID int
}

func New() *Calendar {
	return &Calendar{
		events: make(map[int]*Event),
		nextID: 1,
	}
}

func (c *Calendar) CreateEvent(userID int, dateStr, text string) (int, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, ErrInvalidDate
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextID
	c.nextID++
	c.events[id] = &Event{
		ID:     id,
		UserID: userID,
		Date:   date,
		Text:   text,
	}
	return id, nil
}

func (c *Calendar) UpdateEvent(id, userID int, dateStr, text string) error {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ErrInvalidDate
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	ev, ok := c.events[id]
	if !ok {
		return ErrEventNotFound
	}
	ev.UserID = userID
	ev.Date = date
	ev.Text = text
	return nil
}

func (c *Calendar) DeleteEvent(id int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.events[id]; !ok {
		return ErrEventNotFound
	}
	delete(c.events, id)
	return nil
}

func (c *Calendar) EventsForDay(userID int, dayStr string) ([]Event, error) {
	day, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		return nil, ErrInvalidDate
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	var res []Event
	for _, ev := range c.events {
		if ev.UserID == userID && ev.Date.Equal(day) {
			res = append(res, *ev)
		}
	}
	return res, nil
}

func (c *Calendar) EventsForWeek(userID int, dayStr string) ([]Event, error) {
	day, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		return nil, ErrInvalidDate
	}
	// Найдём понедельник недели
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := day.AddDate(0, 0, -(weekday - 1))
	sunday := monday.AddDate(0, 0, 6)

	c.mu.RLock()
	defer c.mu.RUnlock()
	var res []Event
	for _, ev := range c.events {
		if ev.UserID == userID && !ev.Date.Before(monday) && !ev.Date.After(sunday) {
			res = append(res, *ev)
		}
	}
	return res, nil
}

func (c *Calendar) EventsForMonth(userID int, dayStr string) ([]Event, error) {
	day, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		return nil, ErrInvalidDate
	}
	firstOfMonth := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	c.mu.RLock()
	defer c.mu.RUnlock()
	var res []Event
	for _, ev := range c.events {
		if ev.UserID == userID && !ev.Date.Before(firstOfMonth) && !ev.Date.After(lastOfMonth) {
			res = append(res, *ev)
		}
	}
	return res, nil
}
