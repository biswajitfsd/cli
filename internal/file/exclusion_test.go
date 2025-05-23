package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultExclusions(t *testing.T) {
	separator := string(os.PathSeparator)
	for _, ex := range DefaultExclusions() {
		exParts := strings.Split(ex, separator)
		assert.Greaterf(t, len(exParts), 0, "failed to assert that %s used correct separator. Proper separator %s", ex, separator)
	}
}

func TestExclusionsWithTokenEnvVariable(t *testing.T) {
	oldEnvValue := os.Getenv(debrickedExclusionEnvVar)
	err := os.Setenv(debrickedExclusionEnvVar, "*/**.lock,**/node_modules/**,*\\**.ex")

	if err != nil {
		t.Fatalf("failed to set env var %s", debrickedExclusionEnvVar)
	}

	defer func(key, value string) {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("failed to reset env var %s", debrickedExclusionEnvVar)
		}
	}(debrickedExclusionEnvVar, oldEnvValue)

	gt := []string{"*/**.lock", "**/node_modules/**", "*\\**.ex"}
	exclusions := Exclusions()
	assert.Equal(t, gt, exclusions)

}

func TestExclusionsWithEmptyTokenEnvVariable(t *testing.T) {
	oldEnvValue := os.Getenv(debrickedExclusionEnvVar)
	err := os.Setenv(debrickedExclusionEnvVar, "")

	if err != nil {
		t.Fatalf("failed to set env var %s", debrickedExclusionEnvVar)
	}

	defer func(key, value string) {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("failed to reset env var %s", debrickedExclusionEnvVar)
		}
	}(debrickedExclusionEnvVar, oldEnvValue)

	gt := []string{
		"**/node_modules/**",
		"**/vendor/**",
		"**/.git/**",
		"**/obj/**",
		"**/bower_components/**",
		"**/.vscode-test/**",
	}
	defaultExclusions := Exclusions()
	assert.Equal(t, gt, defaultExclusions)
}

func TestExclude(t *testing.T) {
	var files []string
	_ = filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			files = append(files, path)

			return nil
		})

	cases := []struct {
		name               string
		exclusions         []string
		expectedExclusions []string
	}{
		{
			name:               "NoExclusions",
			exclusions:         []string{},
			expectedExclusions: []string{},
		},
		{
			name:               "InvalidFileExclusion",
			exclusions:         []string{"composer.json"},
			expectedExclusions: []string{},
		},
		{
			name:               "FileExclusionWithDoublestar",
			exclusions:         []string{"**/composer.json"},
			expectedExclusions: []string{"composer.json", "composer.json"}, // Two composer.json files in testdata folder
		},
		{
			name:               "DirectoryExclusion",
			exclusions:         []string{"*/composer/*"},
			expectedExclusions: []string{"composer.json", "composer.lock"},
		},
		{
			name:               "DirectoryExclusionWithRelPath",
			exclusions:         []string{"testdata/go/*"},
			expectedExclusions: []string{"go.mod"},
		},
		{
			name:               "ExtensionExclusionWithWildcardAndDoublestar",
			exclusions:         []string{"**/*.mod"},
			expectedExclusions: []string{"go.mod", "go.mod"}, // Two go.mod files in testdata folder
		},
		{
			name:               "DirectoryExclusionWithDoublestar",
			exclusions:         []string{"**/yarn/**"},
			expectedExclusions: []string{"yarn", "yarn.lock"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var excludedFiles []string
			for _, file := range files {
				if Excluded(c.exclusions, []string{}, file) {
					excludedFiles = append(excludedFiles, file)
				}
			}

			assert.Equal(t, len(c.expectedExclusions), len(excludedFiles), "failed to assert that the same number of files were ignored")

			for _, file := range excludedFiles {
				baseName := filepath.Base(file)
				asserted := false
				for _, expectedExcludedFile := range c.expectedExclusions {
					if baseName == expectedExcludedFile {
						asserted = true

						break
					}
				}

				assert.Truef(t, asserted, "%s ignored when it should pass", file)
			}
		})
	}
}

func TestExcluded(t *testing.T) {
	cases := []struct {
		name       string
		exclusions []string
		inclusions []string
		path       string
		expected   bool
	}{
		{
			name:       "NodeModules",
			exclusions: []string{"**/node_modules/**"},
			inclusions: []string{},
			path:       "node_modules/package.json",
			expected:   true,
		},
		{
			name:       "Inclusions",
			exclusions: []string{"**/node_modules/**"},
			inclusions: []string{"**/package.json"},
			path:       "node_modules/package.json",
			expected:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, Excluded(c.exclusions, c.inclusions, c.path))
		})
	}

}
