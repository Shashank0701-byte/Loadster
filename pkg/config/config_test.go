package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_Valid(t *testing.T) {
	yamlContent := `
target: https://api.example.com
headers:
  Content-Type: application/json
  Authorization: Bearer test-token
timeout: 5s
stages:
  - users: 10
    duration: 30s
  - users: 100
    duration: 1m
`
	cfg, err := Parse(strings.NewReader(yamlContent))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "https://api.example.com", cfg.Target)
	assert.Equal(t, "5s", cfg.Timeout)
	assert.Equal(t, "application/json", cfg.Headers["Content-Type"])
	assert.Equal(t, "Bearer test-token", cfg.Headers["Authorization"])

	require.Len(t, cfg.Stages, 2)
	assert.Equal(t, 10, cfg.Stages[0].Users)
	assert.Equal(t, 30*time.Second, cfg.Stages[0].Duration)
	assert.Equal(t, 100, cfg.Stages[1].Users)
	assert.Equal(t, 1*time.Minute, cfg.Stages[1].Duration)
}

func TestParse_InvalidTarget(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErrSub  string
	}{
		{
			name: "empty target",
			yamlContent: `
stages:
  - users: 10
    duration: 30s
`,
			wantErrSub: "target URL is required",
		},
		{
			name: "invalid url scheme",
			yamlContent: `
target: ftp://api.example.com
stages:
  - users: 10
    duration: 30s
`,
			wantErrSub: "target URL scheme must be http or https",
		},
		{
			name: "malformed url",
			yamlContent: `
target: ://invalid
stages:
  - users: 10
    duration: 30s
`,
			wantErrSub: "invalid target URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tt.yamlContent))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrSub)
		})
	}
}

func TestParse_InvalidStages(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErrSub  string
	}{
		{
			name: "no stages",
			yamlContent: `
target: https://api.example.com
`,
			wantErrSub: "at least one stage must be defined",
		},
		{
			name: "invalid users",
			yamlContent: `
target: https://api.example.com
stages:
  - users: 0
    duration: 30s
`,
			wantErrSub: "users must be greater than 0",
		},
		{
			name: "missing duration",
			yamlContent: `
target: https://api.example.com
stages:
  - users: 10
`,
			wantErrSub: "duration is required",
		},
		{
			name: "invalid duration format",
			yamlContent: `
target: https://api.example.com
stages:
  - users: 10
    duration: 30x
`,
			wantErrSub: "invalid duration format",
		},
		{
			name: "negative duration",
			yamlContent: `
target: https://api.example.com
stages:
  - users: 10
    duration: -5s
`,
			wantErrSub: "duration must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tt.yamlContent))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrSub)
		})
	}
}

func TestParse_InvalidTimeout(t *testing.T) {
	yamlContent := `
target: https://api.example.com
timeout: 5x
stages:
  - users: 10
    duration: 30s
`
	_, err := Parse(strings.NewReader(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid timeout format")
}
