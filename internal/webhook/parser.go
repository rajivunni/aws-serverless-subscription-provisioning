package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type SubscriptionStatus string

const (
	Active    SubscriptionStatus = "ACTIVE"
	Paused                       = "PAUSED"
	Cancelled                    = "CANCELLED"
	Expired                      = "EXPIRED"
)

type OrderItemProperty struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type OrderItem struct {
	Id         int                 `json:"id"`
	ProductId  string              `json:"product_id"`
	Properties []OrderItemProperty `json:"properties"`
}

type Event struct {
	Id                       int                `json:"id"`
	OrderPlacedAt            string             `json:"order_placed"`
	CancellationScheduledFor string             `json:"cancellation_scheduled_for"`
	Email                    string             `json:"email"`
	OrderId                  string             `json:"order_id"`
	CustomerId               string             `json:"customer_id"`
	Status                   SubscriptionStatus `json:"status"`
	Log                      []interface{}      `json:"log"`
	Items                    []OrderItem        `json:"items"`
	Lastname                 string             `json:"last_name"`
	Firstname                string             `json:"first_name"`
}

func parseEventTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time value")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s", raw)
}

// CancellationExpirationTime returns the time at which the subscription's access period ends
// as signalled by the cancellation_scheduled_for field in the subscription webhook payload.
// If the field is missing, we fallback to now so cancelled events can be handled immediately.
func (e *Event) CancellationExpirationTime() (time.Time, error) {
	raw := strings.TrimSpace(e.CancellationScheduledFor)
	if raw == "" {
		return time.Now().UTC(), nil
	}
	return parseEventTime(raw)
}

func (e *Event) CustomerName() string {
	var name string
	if e.Firstname != "" {
		name = e.Firstname
	}
	if e.Lastname != "" {
		name += " " + e.Lastname
	}
	return strings.TrimSpace(name)
}

func (e *Event) Key() string {
	return fmt.Sprintf("subscription_%d_%s_logsl_%d", e.Id, strings.ToLower(string(e.Status)), len(e.Log))
}

var gsmKeys = []string{"GSM Number", "Subscription Mobile Number", "Mobile Number"}

func (e *Event) Msisdn() (string, error) {
	if e.Items == nil || len(e.Items) == 0 {
		log.Printf("no items found in event. skipping provisioning.")
		return "", fmt.Errorf("no items found in event")
	}
	var msisdn string
	for _, item := range e.Items {
		props := item.Properties
		if props == nil || len(props) == 0 {
			continue
		}
		for _, p := range props {
			for _, gsmKey := range gsmKeys {
				if p.Key == gsmKey {
					msisdn = p.Value
					break
				}
			}
		}
	}
	if msisdn == "" {
		return "", fmt.Errorf("no GSM number found in event items")
	}
	return msisdn, nil
}

func Parse(payload []byte) (*Event, error) {
	var event *Event
	err := json.Unmarshal(payload, &event)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event must be a JSON object")
	}
	return event, nil
}
