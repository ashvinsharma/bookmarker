package bookmark

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBookmark(t *testing.T) {
	tests := []struct {
		name       string
		level      int
		indentSize int
		bookmark   *Bookmark
		want       string
	}{
		{
			name:       "Add bookmark with no children",
			level:      0,
			indentSize: 2,
			bookmark: &Bookmark{
				Title: "Google",
				URL:   "https://www.google.com/",
				Tags:  []string{"search", "internet"},
			},
			want: `<DT><A HREF = "https://www.google.com/" ADD_DATE = "" LAST_MODIFIED = "" TAGS = "search,internet">Google</A>
`,
		},
		{
			name:       "Add bookmark with children",
			level:      1,
			indentSize: 2,
			bookmark: &Bookmark{
				Title: "Sublink",
				URL:   "https://www.example.com/sublink",
			},
			want: `  <DT><A HREF = "https://www.example.com/sublink" ADD_DATE = "" LAST_MODIFIED = "">Sublink</A>
`,
		},
		{
			name:       "Add bookmark with nested folder and bookmarks",
			level:      0,
			indentSize: 2,
			bookmark: &Bookmark{
				Title: "Google",
				URL:   "https://www.google.com/",
				Tags:  []string{"search", "internet"},
				Children: []*Bookmark{
					{
						Title: "Sublink",
						URL:   "https://www.example.com/sublink",
					},
					{
						Title: "CMD",
						Children: []*Bookmark{
							{
								Title: "Grafana",
								URL:   "https://www.grafana.com/sublink",
							},
						},
					},
				},
			},
			want: `<DT><A HREF = "https://www.google.com/" ADD_DATE = "" LAST_MODIFIED = "" TAGS = "search,internet">Google</A>
<DL><p>
  <DT><A HREF = "https://www.example.com/sublink" ADD_DATE = "" LAST_MODIFIED = "">Sublink</A>
  <DT><H3>CMD</H3>
  <DL><p>
    <DT><A HREF = "https://www.grafana.com/sublink" ADD_DATE = "" LAST_MODIFIED = "">Grafana</A>
  </DL><p>
</DL><p>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format(tt.level, tt.bookmark, tt.indentSize)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseInputData(t *testing.T) {
	yamlData := `
bookmarks:
  - title: Google
    url: https://www.google.com/
    tags:
      - search
      - internet
  - title: Example
    url: https://www.example.com/
`
	want := &File{
		Bookmarks: []*Bookmark{
			{
				Title: "Google",
				URL:   "https://www.google.com/",
				Tags:  []string{"search", "internet"},
			},
			{
				Title: "Example",
				URL:   "https://www.example.com/",
			},
		},
	}

	got, err := parseInput([]byte(yamlData))
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGenerateBookmarks(t *testing.T) {
	tests := []struct {
		name       string
		inputData  []byte
		wantOutput string
		wantErr    bool
	}{
		{
			name: "Valid Input",
			inputData: []byte(`
bookmarks:
  - title: Google
    url: https://www.google.com/
    tags:
      - search
      - internet
    keyword: keyword
  - title: Amazon
    url: https://www.amazon.com/
    tags:
      - shopping
    keyword: amazon
`),
			wantOutput: `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks Menu</H1>
<DT><H3 ADD_DATE="" LAST_MODIFIED="" PERSONAL_TOOLBAR_FOLDER="true">Bookmarks Toolbar</H3>
<DL><p><DT><A HREF = "https://www.google.com/" ADD_DATE = "" LAST_MODIFIED = "" TAGS = "search,internet" SHORTCUTURL = "keyword">Google</A>
<DT><A HREF = "https://www.amazon.com/" ADD_DATE = "" LAST_MODIFIED = "" TAGS = "shopping" SHORTCUTURL = "amazon">Amazon</A>
</DL><p>`,
			wantErr: false,
		},
		{
			name:       "Invalid Input - Empty string",
			inputData:  []byte(""),
			wantOutput: "",
			wantErr:    true,
		},
		{
			name: "Invalid Input - Malformed YAML",
			inputData: []byte(`
bookmarks:
  - title: Google
    url: https://www.google.com/
    tags:
      - search
      - internet
    keyword: keyword
  - title: Amazon
    url: https://www.amazon.com/
    tags:
      - shopping
    keyword: amazon
  invalid_field: invalid_value
`),
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, _ := os.Pipe()
			// Redirect output to the string builder
			original := os.Stdout
			defer func(r *os.File, w *os.File) {
				_ = r.Close()
				_ = w.Close()
			}(r, w)
			defer func() { os.Stdout = original }()
			os.Stdout = w

			err := generate(tt.inputData)
			if tt.wantErr {
				assert.Error(t, err, "Expected an error, but got none")
				return
			} else {
				if !assert.NoError(t, err, "Unexpected error occurred") {
					return
				}
			}

			_ = w.Close()
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			assert.NoError(t, err, "Unexpected error occurred while copying output to buffer")

			assert.Equal(t, tt.wantOutput, buf.String(), "Output doesn't match expected")
		})
	}
}
