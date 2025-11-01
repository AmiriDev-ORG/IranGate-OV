package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CheckHostNodesResponse struct {
	Nodes map[string]struct {
		ASN      string   `json:"asn"`
		IP       string   `json:"ip"`
		Location []string `json:"location"`
	} `json:"nodes"`
}

var availableNodes []string

func getCheckHostNodes() ([]string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("GET", "https://check-host.net/nodes/ips", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result CheckHostNodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	var nodeList []string
	for nodeName := range result.Nodes {
		nodeList = append(nodeList, nodeName)
	}
	return nodeList, nil
}

type PingCheckRequest struct {
	Host string `json:"host"`
	Mode string `json:"mode"`
}
type PingCheckResponse struct {
	RequestID     string            `json:"request_id"`
	PermanentLink string            `json:"permanent_link"`
	Nodes         map[string]string `json:"nodes"`
	Timestamp     time.Time         `json:"timestamp"`
	Status        string            `json:"status"`
}
type PingResult struct {
	Status    string    `json:"status"`
	RTT       float64   `json:"rtt"`
	Responder string    `json:"responder"`
	Timestamp time.Time `json:"timestamp"`
}
type NodeResult struct {
	NodeName string       `json:"node_name"`
	Location string       `json:"location"`
	Results  []PingResult `json:"results"`
}
type PingCheckStatus struct {
	RequestID     string       `json:"request_id"`
	PermanentLink string       `json:"permanent_link"`
	Status        string       `json:"status"`
	Host          string       `json:"host"`
	Mode          string       `json:"mode"`
	Timestamp     time.Time    `json:"timestamp"`
	Nodes         []NodeResult `json:"nodes"`
	Summary       struct {
		TotalNodes      int     `json:"total_nodes"`
		SuccessfulNodes int     `json:"successful_nodes"`
		FailedNodes     int     `json:"failed_nodes"`
		AverageRTT      float64 `json:"average_rtt"`
	} `json:"summary"`
}
type CheckHostNetResponse struct {
	OK            int                 `json:"ok"`
	RequestID     string              `json:"request_id"`
	PermanentLink string              `json:"permanent_link"`
	Nodes         map[string][]string `json:"nodes"`
}
type CheckHostNetResult map[string]interface{}

func initiatePingCheckHandler(w http.ResponseWriter, r *http.Request) {
	var req PingCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	nodes, err := getCheckHostNodes()
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to get nodes from check-host.net: %v", err), http.StatusInternalServerError)
		return
	}
	if len(nodes) == 0 {
		sendError(w, "No nodes available from check-host.net", http.StatusInternalServerError)
		return
	}
	availableNodes = nodes
	var hostToCheck string
	switch req.Mode {
	case "current_server":
		hostToCheck = getServerExternalIP()
	case "custom_ip":
		hostToCheck = req.Host
	default:
		sendError(w, "Invalid mode. Use 'current_server' or 'custom_ip'", http.StatusBadRequest)
		return
	}
	if hostToCheck == "" {
		sendError(w, "Could not determine host to check", http.StatusBadRequest)
		return
	}
	requestID, permanentLink, err := initiateCheckHostNetPing(hostToCheck)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to initiate ping check: %v", err), http.StatusInternalServerError)
		return
	}
	nodeMap := make(map[string]string)
	for _, node := range nodes {
		nodeMap[node] = getNodeLocation(node)
	}
	response := PingCheckResponse{
		RequestID:     requestID,
		PermanentLink: permanentLink,
		Nodes:         nodeMap,
		Timestamp:     time.Now(),
		Status:        "pending",
	}
	sendSuccess(w, "Ping check initiated successfully", response)
}
func getPingCheckStatusHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getURLParam(r, "request_id")
	if requestID == "" {
		sendError(w, "Request ID is required", http.StatusBadRequest)
		return
	}
	results, err := getCheckHostNetResults(requestID)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to get results: %v", err), http.StatusInternalServerError)
		return
	}
	status := parsePingResults(requestID, results)
	sendSuccess(w, "Ping check status retrieved successfully", status)
}
func getPingCheckHistoryHandler(w http.ResponseWriter, r *http.Request) {
	history := []PingCheckStatus{}
	sendSuccess(w, "Ping check history retrieved successfully", map[string]interface{}{
		"history": history,
		"total":   len(history),
	})
}
func initiateCheckHostNetPing(host string) (string, string, error) {
	if len(availableNodes) == 0 {
		return "", "", fmt.Errorf("no nodes available")
	}
	nodesParam := strings.Join(availableNodes, ",")
	url := fmt.Sprintf("https://check-host.net/check-ping?host=%s&node_count=%d&nodes=%s",
		host, len(availableNodes), nodesParam)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	var result CheckHostNetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	if result.OK != 1 {
		return "", "", fmt.Errorf("check-host.net API returned error")
	}
	return result.RequestID, result.PermanentLink, nil
}
func getCheckHostNetResults(requestID string) (CheckHostNetResult, error) {
	url := fmt.Sprintf("https://check-host.net/check-result/%s", requestID)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("DEBUG: Response body from check-host.net: %s\n", string(body))
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response from check-host.net")
	}
	bodyStr := strings.TrimSpace(string(body))
	if !strings.HasPrefix(bodyStr, "{") && !strings.HasPrefix(bodyStr, "[") {
		return nil, fmt.Errorf("invalid JSON response from check-host.net: %s", bodyStr[:min(len(bodyStr), 100)])
	}
	var result CheckHostNetResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v, body: %s", err, bodyStr[:min(len(bodyStr), 200)])
	}
	return result, nil
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func parsePingResults(requestID string, results CheckHostNetResult) PingCheckStatus {
	status := PingCheckStatus{
		RequestID: requestID,
		Status:    "completed",
		Timestamp: time.Now(),
		Nodes:     []NodeResult{},
	}
	successfulNodes := 0
	failedNodes := 0
	totalRTT := 0.0
	rttCount := 0
	for nodeName, nodeData := range results {
		if nodeData == nil {
			status.Nodes = append(status.Nodes, NodeResult{
				NodeName: nodeName,
				Location: getNodeLocation(nodeName),
				Results:  []PingResult{},
			})
			failedNodes++
			continue
		}
		nodeResult := NodeResult{
			NodeName: nodeName,
			Location: getNodeLocation(nodeName),
			Results:  []PingResult{},
		}
		if resultsSlice, ok := nodeData.([]interface{}); ok {
			for _, resultGroup := range resultsSlice {
				if resultGroupSlice, ok := resultGroup.([]interface{}); ok {
					for _, result := range resultGroupSlice {
						if resultSlice, ok := result.([]interface{}); ok && len(resultSlice) >= 2 {
							pingResult := PingResult{
								Timestamp: time.Now(),
							}
							if statusStr, ok := resultSlice[0].(string); ok {
								switch statusStr {
								case "OK":
									pingResult.Status = "OK"
									if rtt, ok := resultSlice[1].(float64); ok {
										pingResult.RTT = rtt * 1000
										totalRTT += pingResult.RTT
										rttCount++
									}
									if len(resultSlice) > 2 {
										if responder, ok := resultSlice[2].(string); ok {
											pingResult.Responder = responder
										}
									}
								case "TIMEOUT":
									pingResult.Status = "TIMEOUT"
								case "MALFORMED":
									pingResult.Status = "MALFORMED"
								default:
									pingResult.Status = "NO_RESOLUTION"
								}
							}
							nodeResult.Results = append(nodeResult.Results, pingResult)
						}
					}
				}
			}
		}
		status.Nodes = append(status.Nodes, nodeResult)
		if len(nodeResult.Results) > 0 {
			hasSuccess := false
			for _, result := range nodeResult.Results {
				if result.Status == "OK" {
					hasSuccess = true
					break
				}
			}
			if hasSuccess {
				successfulNodes++
			} else {
				failedNodes++
			}
		} else {
			failedNodes++
		}
	}
	status.Summary.TotalNodes = len(status.Nodes)
	status.Summary.SuccessfulNodes = successfulNodes
	status.Summary.FailedNodes = failedNodes
	if rttCount > 0 {
		status.Summary.AverageRTT = totalRTT / float64(rttCount)
	}
	return status
}
func getServerExternalIP() string {
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}
	for _, service := range services {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		ip := strings.TrimSpace(string(body))
		if isValidIP(ip) {
			return ip
		}
	}
	return ""
}
func getNodeLocation(nodeName string) string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", "https://check-host.net/nodes/hosts", nil)
	if err != nil {
		return nodeName
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nodeName
	}
	defer resp.Body.Close()
	var result CheckHostNodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nodeName
	}
	if nodeInfo, exists := result.Nodes[nodeName]; exists && len(nodeInfo.Location) >= 2 {
		return fmt.Sprintf("%s (%s)", nodeInfo.Location[1], nodeInfo.Location[0])
	}
	return nodeName
}
func isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
		if part[0] == '0' && len(part) > 1 {
			return false
		}
	}
	return true
}
