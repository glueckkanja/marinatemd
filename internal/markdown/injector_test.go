package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjector_FindMarkers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
		wantErr  bool
	}{
		{
			name: "single marker",
			content: `# Documentation

Description: <!-- MARINATED: app_config -->

Type: object`,
			expected: []string{"app_config"},
			wantErr:  false,
		},
		{
			name: "multiple markers",
			content: `# Documentation

### Variable 1
Description: <!-- MARINATED: app_config -->

### Variable 2
Description: <!-- MARINATED: database_config -->

### Variable 3
Description: <!-- MARINATED: cache_config -->`,
			expected: []string{"app_config", "database_config", "cache_config"},
			wantErr:  false,
		},
		{
			name: "no markers",
			content: `# Documentation

This file has no MARINATED markers.`,
			expected: []string{},
			wantErr:  false,
		},
		{
			name: "marker with spaces",
			content: `Description: <!-- MARINATED:  app_config  -->`,
			expected: []string{"app_config"},
			wantErr:  false,
		},
		{
			name: "marker with underscore in name",
			content: `Description: <!-- MARINATED: my_complex_variable_name -->`,
			expected: []string{"my_complex_variable_name"},
			wantErr:  false,
		},
		{
			name: "marker with escaped underscore in markdown",
			content: `Description: <!-- MARINATED: app\_config -->`,
			expected: []string{"app_config"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Test FindMarkers
			injector := NewInjector()
			got, err := injector.FindMarkers(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindMarkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.expected) {
				t.Errorf("FindMarkers() got %d markers, want %d", len(got), len(tt.expected))
				return
			}

			for i, marker := range got {
				if marker != tt.expected[i] {
					t.Errorf("FindMarkers()[%d] = %v, want %v", i, marker, tt.expected[i])
				}
			}
		})
	}
}

func TestInjector_InjectIntoFile(t *testing.T) {
	tests := []struct {
		name             string
		originalContent  string
		variableName     string
		markdownContent  string
		expectedContains []string
		wantErr          bool
	}{
		{
			name: "inject after marker preserving prefix",
			originalContent: `### app_config

Description: <!-- MARINATED: app_config -->

Type: object`,
			variableName:    "app_config",
			markdownContent: "- `database` - (Optional) Database configuration\n  - `host` - (Required) Database host\n- `cache` - (Optional) Cache configuration",
			expectedContains: []string{
				"Description: <!-- MARINATED: app_config -->",
				"- `database` - (Optional) Database configuration",
				"- `cache` - (Optional) Cache configuration",
				"<!-- /MARINATED: app_config -->",
				"Type: object",
			},
			wantErr: false,
		},
		{
			name: "inject replaces existing content",
			originalContent: `### app_config

Description: <!-- MARINATED: app_config -->

Some old content that should be replaced.
More old content.

Type: object`,
			variableName:    "app_config",
			markdownContent: "- `new_field` - (Required) New field description",
			expectedContains: []string{
				"Description: <!-- MARINATED: app_config -->",
				"- `new_field` - (Required) New field description",
				"<!-- /MARINATED: app_config -->",
				"Type: object",
			},
			wantErr: false,
		},
		{
			name: "marker not found returns error",
			originalContent: `### some_var

Description: Regular description

Type: string`,
			variableName:    "nonexistent",
			markdownContent: `Some content`,
			wantErr:         true,
		},
		{
			name: "inject with escaped markdown",
			originalContent: `Description: <!-- MARINATED: test_var -->

Type: object`,
			variableName:    "test_var",
			markdownContent: "- `field1` - (Required) A **bold** description\n- `field2` - (Optional) With *italic* text",
			expectedContains: []string{
				"<!-- MARINATED: test_var -->",
				"- `field1` - (Required) A **bold** description",
				"- `field2` - (Optional) With *italic* text",
				"<!-- /MARINATED: test_var -->",
				"Type: object",
			},
			wantErr: false,
		},
		{
			name: "inject with escaped underscore in marker",
			originalContent: `### app\_config

Description: <!-- MARINATED: app\_config -->

Type: object`,
			variableName:    "app_config",
			markdownContent: "- `database` - (Required) Database configuration",
			expectedContains: []string{
				"<!-- MARINATED: app\\_config -->",
				"- `database` - (Required) Database configuration",
				"<!-- /MARINATED: app\\_config -->",
				"Type: object",
			},
			wantErr: false,
		},
		{
			name:             "re-inject updates existing content (idempotency)",
			originalContent:  "### app_config\n\nDescription: <!-- MARINATED: app_config -->\n\n- `old_field` - (Required) Old description\n\n<!-- /MARINATED: app_config -->\n\nType: object",
			variableName:     "app_config",
			markdownContent:  "- `new_field` - (Required) New description\n- `another_field` - (Optional) Another description",
			expectedContains: []string{
				"<!-- MARINATED: app_config -->",
				"- `new_field` - (Required) New description",
				"- `another_field` - (Optional) Another description",
				"<!-- /MARINATED: app_config -->",
				"Type: object",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.originalContent), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Test InjectIntoFile
			injector := NewInjector()
			err := injector.InjectIntoFile(tmpFile, tt.variableName, tt.markdownContent)

			if (err != nil) != tt.wantErr {
				t.Errorf("InjectIntoFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // If we expected an error, we're done
			}

			// Read the file back and verify content
			resultContent, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read result file: %v", err)
			}

			resultStr := string(resultContent)

			// Check that all expected strings are present
			for _, expected := range tt.expectedContains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("InjectIntoFile() result does not contain expected string:\nExpected: %q\nGot: %q", expected, resultStr)
				}
			}

			// Verify some form of the marker is still present (either escaped or unescaped)
			markerUnescaped := "<!-- MARINATED: " + tt.variableName + " -->"
			markerEscaped := "<!-- MARINATED: " + strings.ReplaceAll(tt.variableName, "_", "\\_") + " -->"
			if !strings.Contains(resultStr, markerUnescaped) && !strings.Contains(resultStr, markerEscaped) {
				t.Errorf("InjectIntoFile() removed the marker, but it should be preserved\nLooking for either: %q or %q\nGot: %q", markerUnescaped, markerEscaped, resultStr)
			}
			
			// For the re-injection test, verify old content is gone
			if tt.name == "re-inject updates existing content (idempotency)" {
				if strings.Contains(resultStr, "old_field") {
					t.Errorf("InjectIntoFile() did not remove old content on re-injection\nGot: %q", resultStr)
				}
			}
		})
	}
}

func TestInjector_InjectIntoFile_PreservesStructure(t *testing.T) {
	originalContent := `# Terraform Module

## Inputs

### app_config

Description: <!-- MARINATED: app_config -->

Type: object

Default: n/a

### another_var

Description: Some other variable

Type: string

Default: "default"`

	expectedMarkdown := "- `database` - (Required) Database settings\n  - `host` - (Required) The database host\n  - `port` - (Optional) The database port\n- `cache` - (Optional) Cache settings\n  - `ttl` - (Optional) Time to live"

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(tmpFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Inject markdown
	injector := NewInjector()
	if err := injector.InjectIntoFile(tmpFile, "app_config", expectedMarkdown); err != nil {
		t.Fatalf("InjectIntoFile() failed: %v", err)
	}

	// Read result
	resultContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read result file: %v", err)
	}

	resultStr := string(resultContent)

	// Verify structure is preserved
	checks := []string{
		"# Terraform Module",
		"## Inputs",
		"### app_config",
		"<!-- MARINATED: app_config -->",
		"- `database` - (Required) Database settings",
		"<!-- /MARINATED: app_config -->",
		"Type: object",
		"Default: n/a",
		"### another_var",
		"Description: Some other variable",
		`Default: "default"`,
	}

	for _, check := range checks {
		if !strings.Contains(resultStr, check) {
			t.Errorf("Result missing expected content: %q\nFull result:\n%s", check, resultStr)
		}
	}
}
