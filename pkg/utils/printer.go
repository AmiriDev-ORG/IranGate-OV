package utils
import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/fatih/color"
)
func PrintError(err error) {
	color.Red("Error: %v", err)
}
func PrintSuccess(msg string) {
	color.Green(msg)
}
func PrintInfo(msg string) {
	color.Blue(msg)
}
func PrintJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}
func PrintClientList(clients interface{}) {
	color.Green("\nVPN Clients:")
	data, err := json.MarshalIndent(clients, "", "  ")
	if err != nil {
		PrintError(err)
		return
	}
	fmt.Println(string(data))
}
func PrintConfig(config interface{}) {
	color.Green("\nOpenVPN Configuration:")
	configJSON, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println(string(configJSON))
}
func PrintStats(stats interface{}) {
	color.Green("\nOpenVPN Statistics:")
	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(statsJSON))
}
func PrintJobs(jobs interface{}) {
	color.Green("\nScheduled Jobs:")
	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		PrintError(err)
		return
	}
	fmt.Println(string(data))
}
func PrintClientDetails(client interface{}) {
	color.Green("\nClient Details:")
	data, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		PrintError(err)
		return
	}
	fmt.Println(string(data))
}