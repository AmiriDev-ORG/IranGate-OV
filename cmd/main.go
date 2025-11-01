package main
import (
	"fmt"
	"os"
	"github.com/amiridev-org/irangate-ov/pkg/cmd"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
)
func main() {
	utils.InitLogger()
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}