package check

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckJSONResult_JSONResponse(t *testing.T) {
	result := CheckJSONResult{
		Passed: true,
		Total:  3,
		Failed: 0,
		Checks: []CheckJSONItem{
			{Name: "format", Category: "code_quality", Passed: true, DurationMs: 340},
			{Name: "lint", Category: "code_quality", Passed: true, DurationMs: 1200},
			{Name: "secrets", Category: "security", Passed: true, DurationMs: 89},
		},
	}

	data := result.JSONResponse()
	jsonResult, ok := data.(CheckJSONResult)
	require.True(t, ok)
	assert.Equal(t, 3, jsonResult.Total)
	assert.True(t, jsonResult.Passed)
	assert.Len(t, jsonResult.Checks, 3)
}

func TestCheckJSONResult_Failed(t *testing.T) {
	result := CheckJSONResult{
		Passed: false,
		Total:  2,
		Failed: 1,
		Checks: []CheckJSONItem{
			{Name: "format", Category: "code_quality", Passed: true, DurationMs: 340},
			{Name: "lint", Category: "code_quality", Passed: false, DurationMs: 500, Error: "lint issues found"},
		},
	}

	assert.False(t, result.Passed)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, "lint issues found", result.Checks[1].Error)
}

func TestToJSONResult(t *testing.T) {
	results := []allCheckResult{
		{name: "format", category: "code_quality", passed: true, duration: 340 * time.Millisecond},
		{name: "lint", category: "code_quality", passed: false, duration: 500 * time.Millisecond, err: errors.New("lint failed")},
	}

	jsonResult := toJSONResult(results, 1, 1)

	assert.False(t, jsonResult.Passed)
	assert.Equal(t, 2, jsonResult.Total)
	assert.Equal(t, 1, jsonResult.Failed)
	assert.Len(t, jsonResult.Checks, 2)
	assert.Equal(t, "format", jsonResult.Checks[0].Name)
	assert.True(t, jsonResult.Checks[0].Passed)
	assert.Equal(t, int64(340), jsonResult.Checks[0].DurationMs)
	assert.Empty(t, jsonResult.Checks[0].Error)
	assert.False(t, jsonResult.Checks[1].Passed)
	assert.Equal(t, "lint failed", jsonResult.Checks[1].Error)
}

func TestToJSONResult_AllPassed(t *testing.T) {
	results := []allCheckResult{
		{name: "format", category: "code_quality", passed: true, duration: 100 * time.Millisecond},
	}

	jsonResult := toJSONResult(results, 1, 0)
	assert.True(t, jsonResult.Passed)
	assert.Equal(t, 0, jsonResult.Failed)
}
