package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseCPUToMillis(t *testing.T) {
	assert.Equal(t, 50, parseCPUToMillis("50m"))
	assert.Equal(t, 1500, parseCPUToMillis("1.5"))
	assert.Equal(t, 1000, parseCPUToMillis("1"))
	assert.Equal(t, minCPUMillis, parseCPUToMillis(""))
	assert.Equal(t, minCPUMillis, parseCPUToMillis("invalid"))
}

func TestParseMemoryToMi(t *testing.T) {
	assert.Equal(t, 64, parseMemoryToMi("64Mi"))
	assert.Equal(t, 1024, parseMemoryToMi("1Gi"))
	assert.Equal(t, 2, parseMemoryToMi("2048Ki"))
	assert.Equal(t, 64, parseMemoryToMi("67108864")) // 64Mi in bytes
	assert.Equal(t, minMemoryMi, parseMemoryToMi(""))
	assert.Equal(t, minMemoryMi, parseMemoryToMi("invalid"))
}

func TestRunRightSize_DryRunAndApply(t *testing.T) {
	tempDir := t.TempDir()

	initialYAML := `
service:
  targetPort: 8081
resources:
  requests:
    cpu: 1000m
    memory: 512Mi
  limits:
    memory: 1024Mi
`
	testFile := filepath.Join(tempDir, "identity-service.yaml")
	err := os.WriteFile(testFile, []byte(initialYAML), 0644)
	require.NoError(t, err)

	// Test dry run mode
	err = runRightSize(tempDir, "test-namespace", true)
	require.NoError(t, err)

	// Ensure file was NOT modified in dry-run
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, initialYAML, string(content))

	// Test applied mode
	err = runRightSize(tempDir, "test-namespace", false)
	require.NoError(t, err)

	// Ensure file WAS modified with right-sized values
	updatedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	var doc map[string]interface{}
	err = yaml.Unmarshal(updatedContent, &doc)
	require.NoError(t, err)

	res := doc["resources"].(map[string]interface{})
	reqs := res["requests"].(map[string]interface{})
	limits := res["limits"].(map[string]interface{})

	assert.Equal(t, "54m", reqs["cpu"])
	assert.Equal(t, "76Mi", reqs["memory"])
	assert.Equal(t, "114Mi", limits["memory"])
}
