package main

import (
	"fmt"
	_ "modernc.org/sqlite"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "vpn-tool"}

	rootCmd.AddCommand(Setup(), Add(), Delete(), Get(), GetAllUsers(), Server())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
