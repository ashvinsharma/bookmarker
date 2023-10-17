package main

import (
	"fmt"
	"github.com/ashvinsharma/bookmarker/bookmark"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bookmark-maker",
	Short: "Bookmark Maker CLI is a tool to generate bookmarks in a format importable by browsers.",
	Long: `Bookmark Maker CLI is a tool to generate a bookmarks file in a format that can be imported by modern browsers. 
The input file should be in YAML format and must contain the hierarchy of bookmarks and folders. 
Each bookmark entry should have a title, URL, and optional tags and keyword. 
The generated output will be printed to stdout, which you can redirect to a file if needed.`,
	Example: `  bookmark-maker input.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		err := bookmark.Generate(inputFile)
		if err != nil {
			return err
		}
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
