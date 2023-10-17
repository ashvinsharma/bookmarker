package bookmark

import (
	"embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"
	"text/template"
)

//go:embed template.gohtml
var goTemplate embed.FS

type Bookmark struct {
	Title    string   `yaml:"title,omitempty"`
	URL      string   `yaml:"url,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
	Keyword  string   `yaml:"keyword,omitempty"`
	Children []*Bookmark
}

type File struct {
	Bookmarks []*Bookmark `yaml:"bookmarks"`
}

func readInput(inputFile string) ([]byte, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return io.ReadAll(os.Stdin)
	} else {
		if inputFile == "" {
			return nil, fmt.Errorf("input file is required, use -i or pipe the input YAML data")
		}
		return os.ReadFile(inputFile)
	}
}

func parseInput(inputData []byte) (*File, error) {
	var bookmarkFile File
	err := yaml.Unmarshal(inputData, &bookmarkFile)
	if err != nil {
		return nil, err
	}
	return &bookmarkFile, nil
}

// format formats a single bookmark and its children using a template.
func format(level int, b *Bookmark, indentSize int) string {
	indent := strings.Repeat(" ", level*indentSize)

	var sb strings.Builder

	templateFuncs := template.FuncMap{
		"join":     strings.Join,
		"format":   format,
		"isFolder": func(b *Bookmark) bool { return b.URL == "" },
	}
	templateWithFuncs := template.New("template.gohtml").Funcs(templateFuncs)
	t := template.Must(templateWithFuncs.ParseFS(goTemplate, "template.gohtml"))

	err := t.Execute(&sb, map[string]interface{}{
		"Indent":     indent,
		"IndentSize": indentSize,
		"Level":      level + 1,
		"Bookmark":   b,
	})

	if err != nil {
		panic(err)
	}

	return sb.String()
}

// Generate generates a bookmarks file in a format importable by browsers.
// The function takes the input YAML data, processes it, and prints the output to stdout.
func Generate(inputFile string) error {
	inputData, err := readInput(inputFile)
	if err != nil {
		return err
	}

	return generate(inputData)
}

func generate(inputData []byte) error {
	if len(inputData) == 0 {
		return fmt.Errorf("blank data is not allowed")
	}

	bookmarkFile, err := parseInput(inputData)
	if err != nil {
		return err
	}

	bookmarks := bookmarkFile.Bookmarks

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<meta http-equiv="Content-Security-Policy"
      content="default-src 'self'; script-src 'none'; img-src data: *; object-src 'none'"></meta>
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks Menu</H1>
<DL><p>`)

	for _, b := range bookmarks {
		sb.WriteString(format(0, b, 2))
	}

	sb.WriteString("</DL><p>")

	fmt.Print(sb.String())

	return nil
}
