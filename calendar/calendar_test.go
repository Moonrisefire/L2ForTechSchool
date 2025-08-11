package calendar

import (
	"testing"
)

func TestCreateAndGetEvents(t *testing.T) {
	c := New()
	id, err := c.CreateEvent(1, "2023-08-11", "Test Event")
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}
	if id == 0 {
		t.Fatal("Expected non-zero event ID")
	}

	events, err := c.EventsForDay(1, "2023-08-11")
	if err != nil {
		t.Fatalf("EventsForDay failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Text != "Test Event" {
		t.Fatalf("Unexpected event text: %s", events[0].Text)
	}
}

func TestUpdateEvent(t *testing.T) {
	c := New()
	id, _ := c.CreateEvent(1, "2023-08-11", "Old Event")

	err := c.UpdateEvent(id, 1, "2023-08-12", "Updated Event")
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}

	events, _ := c.EventsForDay(1, "2023-08-12")
	if len(events) != 1 || events[0].Text != "Updated Event" {
		t.Fatal("UpdateEvent did not update correctly")
	}
}

func TestDeleteEvent(t *testing.T) {
	c := New()
	id, _ := c.CreateEvent(1, "2023-08-11", "Event to delete")

	err := c.DeleteEvent(id)
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	events, _ := c.EventsForDay(1, "2023-08-11")
	if len(events) != 0 {
		t.Fatal("DeleteEvent did not remove event")
	}
}

func TestInvalidDate(t *testing.T) {
	c := New()
	_, err := c.CreateEvent(1, "invalid-date", "Bad date")
	if err != ErrInvalidDate {
		t.Fatalf("Expected ErrInvalidDate, got %v", err)
	}
}
