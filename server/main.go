package main

import (
	"fmt"
	"os"

	_ "modernc.org/sqlite"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "vpn-tool"}

	rootCmd.AddCommand(Setup(), Add(), Delete(), Get(), GetAllUsers(), Server(), UpdateEndpoints(), Info())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
