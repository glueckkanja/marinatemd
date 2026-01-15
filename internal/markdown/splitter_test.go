package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitter_ExtractSections(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedCount  int
		expectedVars   []string
		wantErr        bool
	}{
		{
			name: "single MARINATED variable",
			content: `# Documentation

## Inputs

### app\_config

Description: <!-- MARINATED: app_config -->
Some content here
<!-- /MARINATED: app_config -->

Type: object

Default: n/a
`,
			expectedCount: 1,
			expectedVars:  []string{"app_config"},
			wantErr:       false,
		},
		{
			name: "multiple MARINATED variables",
			content: `# Documentation

## Inputs

### app\_config

Description: <!-- MARINATED: app_config -->
App config content
<!-- /MARINATED: app_config -->

Type: object
Default: n/a

### database\_config

Description: <!-- MARINATED: database_config -->
Database config content
<!-- /MARINATED: database_config -->

Type: object
Default: n/a

### cache\_settings

Description: <!-- MARINATED: cache_settings -->
Cache settings content
<!-- /MARINATED: cache_settings -->

Type: string
Default: "default"
`,
			expectedCount: 3,
			expectedVars:  []string{"app_config", "database_config", "cache_settings"},
			wantErr:       false,
		},
		{
			name: "mixed MARINATED and non-MARINATED variables",
			content: `# Documentation

## Inputs

### app\_config

Description: <!-- MARINATED: app_config -->
App config content
<!-- /MARINATED: app_config -->

Type: object

### regular\_var

Description: This is not a marinated variable

Type: string

### database\_config

Description: <!-- MARINATED: database_config -->
Database config content
<!-- /MARINATED: database_config -->

Type: object
`,
			expectedCount: 2,
			expectedVars:  []string{"app_config", "database_config"},
			wantErr:       false,
		},
		{
			name: "no MARINATED variables",
			content: `# Documentation

## Inputs

### regular\_var

Description: This is not a marinated variable

Type: string
`,
			expectedCount: 0,
			expectedVars:  []string{},
			wantErr:       false,
		},
		{
			name: "MARINATED marker without escaped underscores",
			content: `# Documentation

## Inputs

### app_config

Description: <!-- MARINATED: app_config -->
App config content
<!-- /MARINATED: app_config -->

Type: object
`,
			expectedCount: 1,
			expectedVars:  []string{"app_config"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter := NewSplitter()
			sections, err := splitter.extractSectionsFromContent(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(sections) != tt.expectedCount {
				t.Errorf("ExtractSections() got %d sections, want %d", len(sections), tt.expectedCount)
			}

			// Check variable names
			for i, expectedVar := range tt.expectedVars {
				if i >= len(sections) {
					t.Errorf("Missing section for variable %s", expectedVar)
					continue
				}
				if sections[i].VariableName != expectedVar {
					t.Errorf("Section[%d] variable name = %s, want %s", i, sections[i].VariableName, expectedVar)
				}
			}

			// Verify content is captured
			for _, section := range sections {
				if section.Content == "" {
					t.Errorf("Section for %s has empty content", section.VariableName)
				}
				// Should contain the heading
				if !strings.Contains(section.Content, "###") {
					t.Errorf("Section for %s missing heading", section.VariableName)
				}
				// Should contain the MARINATED marker
				if !strings.Contains(section.Content, "<!-- MARINATED:") {
					t.Errorf("Section for %s missing MARINATED marker", section.VariableName)
				}
			}
		})
	}
}

func TestSplitter_ExtractSectionsFromFile(t *testing.T) {
	// Create a temporary file with test content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	content := `# Test Documentation

## Inputs

### app\_config

Description: <!-- MARINATED: app_config -->
- Some attribute description
<!-- /MARINATED: app_config -->

Type: object({})
Default: n/a

### database\_settings

Description: <!-- MARINATED: database_settings -->
- Database connection settings
<!-- /MARINATED: database_settings -->

Type: object({})
Default: null
`

	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	splitter := NewSplitter()
	sections, err := splitter.ExtractSections(testFile)
	if err != nil {
		t.Fatalf("ExtractSections() error = %v", err)
	}

	if len(sections) != 2 {
		t.Errorf("ExtractSections() got %d sections, want 2", len(sections))
	}

	expectedVars := []string{"app_config", "database_settings"}
	for i, expected := range expectedVars {
		if sections[i].VariableName != expected {
			t.Errorf("Section[%d] variable name = %s, want %s", i, sections[i].VariableName, expected)
		}
	}
}

func TestSplitter_WriteSection(t *testing.T) {
	tests := []struct {
		name         string
		section      VariableSection
		header       string
		footer       string
		wantContains []string
	}{
		{
			name: "basic section without header/footer",
			section: VariableSection{
				VariableName: "test_var",
				Content:      "### test_var\n\nDescription: Test content\n\nType: string",
			},
			wantContains: []string{"### test_var", "Description: Test content", "Type: string"},
		},
		{
			name: "section with header",
			section: VariableSection{
				VariableName: "test_var",
				Content:      "### test_var\n\nDescription: Test content",
			},
			header:       "# Header Title\n\nThis is a header.",
			wantContains: []string{"# Header Title", "This is a header", "### test_var", "Description: Test content"},
		},
		{
			name: "section with footer",
			section: VariableSection{
				VariableName: "test_var",
				Content:      "### test_var\n\nDescription: Test content",
			},
			footer:       "---\n\nGenerated by marinatemd",
			wantContains: []string{"### test_var", "Description: Test content", "---", "Generated by marinatemd"},
		},
		{
			name: "section with both header and footer",
			section: VariableSection{
				VariableName: "test_var",
				Content:      "### test_var\n\nDescription: Test content",
			},
			header:       "# Header\n",
			footer:       "\n---\nFooter",
			wantContains: []string{"# Header", "### test_var", "Description: Test content", "---", "Footer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.md")

			splitter := NewSplitter()
			if tt.header != "" {
				splitter.SetHeader(tt.header)
			}
			if tt.footer != "" {
				splitter.SetFooter(tt.footer)
			}

			err := splitter.WriteSection(outputPath, tt.section)
			if err != nil {
				t.Fatalf("WriteSection() error = %v", err)
			}

			// Read the written file
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			contentStr := string(content)

			// Check that all expected strings are present
			for _, want := range tt.wantContains {
				if !strings.Contains(contentStr, want) {
					t.Errorf("WriteSection() output missing expected content: %q", want)
				}
			}

			// Verify order (header should come before content, content before footer)
			if tt.header != "" {
				headerIdx := strings.Index(contentStr, tt.header)
				contentIdx := strings.Index(contentStr, tt.section.Content[:10]) // Use first 10 chars
				if headerIdx > contentIdx {
					t.Errorf("Header should appear before content")
				}
			}

			if tt.footer != "" {
				footerIdx := strings.Index(contentStr, tt.footer)
				contentIdx := strings.Index(contentStr, tt.section.Content[:10])
				if footerIdx < contentIdx {
					t.Errorf("Footer should appear after content")
				}
			}
		})
	}
}

func TestSplitter_SplitToFiles(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.md")
	outputDir := filepath.Join(tmpDir, "output")

	content := `# Documentation

## Inputs

### app\_config

Description: <!-- MARINATED: app_config -->
- database - Database settings
- cache - Cache settings
<!-- /MARINATED: app_config -->

Type: object({})
Default: n/a

### storage\_config

Description: <!-- MARINATED: storage_config -->
- bucket - Storage bucket name
- region - Storage region
<!-- /MARINATED: storage_config -->

Type: object({})
Default: null
`

	if err := os.WriteFile(inputFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	splitter := NewSplitter()
	createdFiles, err := splitter.SplitToFiles(inputFile, outputDir)
	if err != nil {
		t.Fatalf("SplitToFiles() error = %v", err)
	}

	if len(createdFiles) != 2 {
		t.Errorf("SplitToFiles() created %d files, want 2", len(createdFiles))
	}

	// Check that files were created with correct names
	expectedFiles := map[string]bool{
		"app_config.md":     false,
		"storage_config.md": false,
	}

	for _, filePath := range createdFiles {
		filename := filepath.Base(filePath)
		if _, exists := expectedFiles[filename]; exists {
			expectedFiles[filename] = true

			// Verify the file exists and has content
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Errorf("Failed to read created file %s: %v", filePath, err)
				continue
			}

			if len(content) == 0 {
				t.Errorf("Created file %s is empty", filePath)
			}

			// Verify content contains the variable section
			contentStr := string(content)
			if !strings.Contains(contentStr, "<!-- MARINATED:") {
				t.Errorf("File %s missing MARINATED marker", filename)
			}
		}
	}

	// Check that all expected files were created
	for filename, created := range expectedFiles {
		if !created {
			t.Errorf("Expected file %s was not created", filename)
		}
	}
}

func TestSplitter_NewSplitterWithTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	headerFile := filepath.Join(tmpDir, "header.md")
	footerFile := filepath.Join(tmpDir, "footer.md")

	headerContent := "# Header\n\nThis is a header."
	footerContent := "---\n\nGenerated by marinatemd"

	if err := os.WriteFile(headerFile, []byte(headerContent), 0600); err != nil {
		t.Fatalf("Failed to create header file: %v", err)
	}
	if err := os.WriteFile(footerFile, []byte(footerContent), 0600); err != nil {
		t.Fatalf("Failed to create footer file: %v", err)
	}

	t.Run("with header and footer", func(t *testing.T) {
		splitter, err := NewSplitterWithTemplate(headerFile, footerFile)
		if err != nil {
			t.Fatalf("NewSplitterWithTemplate() error = %v", err)
		}

		if splitter.headerContent != headerContent {
			t.Errorf("Header content not loaded correctly")
		}
		if splitter.footerContent != footerContent {
			t.Errorf("Footer content not loaded correctly")
		}
	})

	t.Run("with only header", func(t *testing.T) {
		splitter, err := NewSplitterWithTemplate(headerFile, "")
		if err != nil {
			t.Fatalf("NewSplitterWithTemplate() error = %v", err)
		}

		if splitter.headerContent != headerContent {
			t.Errorf("Header content not loaded correctly")
		}
		if splitter.footerContent != "" {
			t.Errorf("Footer should be empty")
		}
	})

	t.Run("with nonexistent header file", func(t *testing.T) {
		_, err := NewSplitterWithTemplate("/nonexistent/header.md", "")
		if err == nil {
			t.Errorf("Expected error for nonexistent header file")
		}
	})
}

func TestSplitter_ComplexMarkdown(t *testing.T) {
	// Test with actual example content structure
	content := `# Terraform Configuration Documentation

<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD013 -->
## Inputs

The following input variables are supported:

### app\_config

Description: <!-- MARINATED: app_config -->

- cache - (Optional) # TODO: Add description for cache
  - redis_url - (Required) # TODO: Add description for redis_url
  - ttl - (Optional) # TODO: Add description for ttl
- database - (Optional) # TODO: Add description for database
  - host - (Required) # TODO: Add description for host
  - port - (Optional) # TODO: Add description for port

<!-- /MARINATED: app_config -->
Type:

` + "```hcl" + `
object({
    database = optional(object({
      host     = string
      port     = optional(number, 5432)
    }))
})
` + "```" + `

Default: n/a

### access\_tier

Description: (Optional) Defines the access tier for BlobStorage.

Type: string

Default: "Hot"
`

	splitter := NewSplitter()
	sections, err := splitter.extractSectionsFromContent(content)
	if err != nil {
		t.Fatalf("extractSectionsFromContent() error = %v", err)
	}

	if len(sections) != 1 {
		t.Errorf("Expected 1 MARINATED section, got %d", len(sections))
	}

	if sections[0].VariableName != "app_config" {
		t.Errorf("Variable name = %s, want app_config", sections[0].VariableName)
	}

	// Verify the content includes all parts
	contentStr := sections[0].Content
	expectedParts := []string{
		"### app\\_config",
		"<!-- MARINATED: app_config -->",
		"<!-- /MARINATED: app_config -->",
		"Type:",
		"```hcl",
		"Default: n/a",
	}

	for _, part := range expectedParts {
		if !strings.Contains(contentStr, part) {
			t.Errorf("Section content missing expected part: %q", part)
		}
	}

	// Should NOT include the next variable
	if strings.Contains(contentStr, "access_tier") {
		t.Errorf("Section should not include content from next variable")
	}
}
