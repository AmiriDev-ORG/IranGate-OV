package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DocumentationEntry struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Content     map[string]interface{} `json:"content"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func getDocumentationHandler(w http.ResponseWriter, r *http.Request) {
	name := getURLParam(r, "name")
	if name == "" {
		sendError(w, "Documentation name is required", http.StatusBadRequest)
		return
	}
	doc, err := getDocumentation(name)
	if err != nil {
		sendError(w, fmt.Sprintf("Documentation not found: %v", err), http.StatusNotFound)
		return
	}
	sendSuccess(w, "Documentation retrieved successfully", doc)
}
func saveDocumentationHandler(w http.ResponseWriter, r *http.Request) {
	var doc DocumentationEntry
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if doc.Name == "" {
		sendError(w, "Documentation name is required", http.StatusBadRequest)
		return
	}
	doc.UpdatedAt = time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}
	err := saveDocumentation(doc)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to save documentation: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Documentation saved successfully", doc)
}
func listDocumentationHandler(w http.ResponseWriter, r *http.Request) {
	docs, err := listAllDocumentation()
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to list documentation: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Documentation list retrieved successfully", map[string]interface{}{
		"docs":  docs,
		"count": len(docs),
	})
}
func getDocumentation(name string) (*DocumentationEntry, error) {
	if name == "check-host.net" {
		return &DocumentationEntry{
			ID:          "check-host-net",
			Name:        "check-host.net",
			Title:       "Check-Host.net Ping API",
			Description: "External ping monitoring service for server connectivity checks",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Content: map[string]interface{}{
				"ping_check_example": `curl -H "Accept: application/json" \
  'https://check-host.net/check-ping?host=example.com&node_count=1&nodes=us1.node.check-host.net'`,
				"ping_check_response": map[string]interface{}{
					"ok":             1,
					"request_id":     "806df9",
					"permanent_link": "https://check-host.net/check-ping/806df9",
					"nodes": map[string][]string{
						"us1.node.check-host.net": {"us", "USA", "Los Angeles", "5.253.30.82", "AS18978"},
						"ch1.node.check-host.net": {"ch", "Switzerland", "Zurich", "179.43.148.195", "AS50837"},
						"pt1.node.check-host.net": {"pt", "Portugal", "Viana", "185.83.213.25", "AS44222"},
					},
				},
				"ping_results_example": `curl -H "Accept: application/json" \
  https://check-host.net/check-result/[request_id]`,
				"ping_results_response": map[string]interface{}{
					"us1.node.check-host.net": [][]interface{}{
						{"OK", 0.044, "94.242.206.94"},
						{"TIMEOUT", 3.005},
						{"MALFORMED", 0.045},
						{"OK", 0.0433},
					},
					"ch1.node.check-host.net": [][]interface{}{
						{nil},
					},
					"pt1.node.check-host.net": nil,
				},
				"results_interpretation": map[string]string{
					"OK":        "Successful ping with response time in seconds",
					"TIMEOUT":   "No response within timeout period (usually 3 seconds)",
					"MALFORMED": "Invalid response received from target",
					"null":      "Node unavailable or error occurred",
				},
				"usage_notes": []string{
					"Each node result contains an array of ping attempts",
					"First element is status (OK, TIMEOUT, MALFORMED)",
					"Second element is response time in seconds (if OK)",
					"Third element is responder IP (if OK)",
					"Multiple attempts may be shown for comprehensive testing",
				},
			},
		}, nil
	}
	return nil, fmt.Errorf("documentation not found: %s", name)
}
func saveDocumentation(_ DocumentationEntry) error {
	return nil
}
func listAllDocumentation() ([]DocumentationEntry, error) {
	docs := []DocumentationEntry{
		{
			ID:          "check-host-net",
			Name:        "check-host.net",
			Title:       "Check-Host.net Ping API",
			Description: "External ping monitoring service for server connectivity checks",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	return docs, nil
}
