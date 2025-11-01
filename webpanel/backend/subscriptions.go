package main
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)
var SubscriptionsDBPath = func() string {
	if dir := os.Getenv("IRANGATE_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "webpanel", "subscriptions.json")
	}
	return "/opt/irangate/webpanel/subscriptions.json"
}()
type Subscription struct {
	ID            string    `json:"id"`
	ClientName    string    `json:"client_name"`
	GroupID       string    `json:"group_id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Status        string    `json:"status"`
	AutoRenew     bool      `json:"auto_renew"`
	PaymentMethod string    `json:"payment_method"`
	TotalPaid     float64   `json:"total_paid"`
	LastPayment   time.Time `json:"last_payment"`
	NextBilling   time.Time `json:"next_billing"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
func loadSubscriptions() ([]Subscription, error) {
	if _, err := os.Stat(SubscriptionsDBPath); os.IsNotExist(err) {
		subs := []Subscription{}
		saveSubscriptions(subs)
		return subs, nil
	}
	data, err := os.ReadFile(SubscriptionsDBPath)
	if err != nil {
		return nil, err
	}
	var subs []Subscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}
func saveSubscriptions(subs []Subscription) error {
	os.MkdirAll(filepath.Dir(SubscriptionsDBPath), 0755)
	data, err := json.MarshalIndent(subs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(SubscriptionsDBPath, data, 0600)
}
func createSubscription(sub Subscription) error {
	subs, err := loadSubscriptions()
	if err != nil {
		return err
	}
	for _, s := range subs {
		if s.ClientName == sub.ClientName {
			return fmt.Errorf("subscription already exists for client")
		}
	}
	sub.ID = fmt.Sprintf("sub_%d", time.Now().Unix())
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	sub.Status = "active"
	subs = append(subs, sub)
	return saveSubscriptions(subs)
}
func updateSubscription(clientName string, updates map[string]interface{}) error {
	subs, err := loadSubscriptions()
	if err != nil {
		return err
	}
	found := false
	for i, sub := range subs {
		if sub.ClientName == clientName {
			if groupID, ok := updates["group_id"].(string); ok {
				subs[i].GroupID = groupID
			}
			if endDate, ok := updates["end_date"].(string); ok {
				if t, err := time.Parse(time.RFC3339, endDate); err == nil {
					subs[i].EndDate = t
				}
			}
			if status, ok := updates["status"].(string); ok {
				subs[i].Status = status
			}
			if autoRenew, ok := updates["auto_renew"].(bool); ok {
				subs[i].AutoRenew = autoRenew
			}
			if notes, ok := updates["notes"].(string); ok {
				subs[i].Notes = notes
			}
			subs[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("subscription not found")
	}
	return saveSubscriptions(subs)
}
func getExpiringSubscriptions(days int) ([]Subscription, error) {
	subs, err := loadSubscriptions()
	if err != nil {
		return nil, err
	}
	threshold := time.Now().AddDate(0, 0, days)
	var expiring []Subscription
	for _, sub := range subs {
		if sub.Status == "active" && sub.EndDate.Before(threshold) && sub.EndDate.After(time.Now()) {
			expiring = append(expiring, sub)
		}
	}
	return expiring, nil
}
func renewSubscription(clientName string, days int) error {
	subs, err := loadSubscriptions()
	if err != nil {
		return err
	}
	for i, sub := range subs {
		if sub.ClientName == clientName {
			subs[i].EndDate = time.Now().AddDate(0, 0, days)
			subs[i].Status = "active"
			subs[i].LastPayment = time.Now()
			subs[i].NextBilling = time.Now().AddDate(0, 0, days)
			subs[i].UpdatedAt = time.Now()
			return saveSubscriptions(subs)
		}
	}
	return fmt.Errorf("subscription not found")
}
func deleteSubscription(clientName string) error {
	subs, err := loadSubscriptions()
	if err != nil {
		return err
	}
	var updated []Subscription
	for _, sub := range subs {
		if sub.ClientName != clientName {
			updated = append(updated, sub)
		}
	}
	return saveSubscriptions(updated)
}