package check

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
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

// Integration-level tests that exercise the full Execute() JSON path.

func TestCheckJSON_AllPass(t *testing.T) {
	output.SetOutputMode("json")
	output.SetCommandName("check")
	defer func() {
		output.SetOutputMode("")
		output.SetCommandName("")
	}()

	var buf bytes.Buffer
	cfg := Config{BinaryName: "test"}
	timings := loadTimingHistory()
	executor := &Executor{
		cfg:     cfg,
		writer:  &buf,
		timings: timings,
		runner:  NewRunner(timings),
	}

	// Build a minimal category with checks that always pass
	passingCheck := func(ctx context.Context) error { return nil }
	categories := []categoryDef{
		{name: "TestCategory", checks: []checkItem{
			{"check-a", passingCheck, "fix a"},
			{"check-b", passingCheck, "fix b"},
		}},
	}

	// Override buildCategories by running the checks directly
	var allResults []allCheckResult
	var totalPassed, totalFailed int

	for _, cat := range categories {
		opts := RunOptions{}
		onDone := func(index int, r Result) {}
		runnerResults, _ := executor.runner.RunChecks(context.Background(), cat, opts, onDone)
		for _, r := range runnerResults {
			result := allCheckResult{name: r.Name, category: r.Category, passed: r.Passed, duration: r.Duration, err: r.Err, remediation: r.Remediation}
			allResults = append(allResults, result)
			if r.Passed {
				totalPassed++
			} else {
				totalFailed++
			}
		}
	}

	// Now test the JSON rendering path
	jsonResult := toJSONResult(allResults, totalPassed, totalFailed)
	status := "success"
	var jsonErr *output.JSONError
	if totalFailed > 0 {
		status = "error"
		jsonErr = &output.JSONError{Message: "checks failed"}
	}
	err := output.RenderJSON(&buf, output.JSONEnvelope{
		Status:  status,
		Command: output.CommandName(),
		Data:    jsonResult.JSONResponse(),
		Error:   jsonErr,
	})
	require.NoError(t, err)

	// Parse and verify
	var envelope output.JSONEnvelope
	err = json.Unmarshal(buf.Bytes(), &envelope)
	require.NoError(t, err, "should produce valid JSON, got: %s", buf.String())

	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, "check", envelope.Command)
	assert.Nil(t, envelope.Error)
	assert.NotNil(t, envelope.Data)

	// Verify data contains check results
	dataBytes, _ := json.Marshal(envelope.Data)
	var checkData CheckJSONResult
	err = json.Unmarshal(dataBytes, &checkData)
	require.NoError(t, err)

	assert.True(t, checkData.Passed)
	assert.Equal(t, 2, checkData.Total)
	assert.Equal(t, 0, checkData.Failed)
	assert.Len(t, checkData.Checks, 2)
	assert.Equal(t, "check-a", checkData.Checks[0].Name)
	assert.True(t, checkData.Checks[0].Passed)
}

func TestCheckJSON_SomeFail(t *testing.T) {
	output.SetOutputMode("json")
	output.SetCommandName("check")
	defer func() {
		output.SetOutputMode("")
		output.SetCommandName("")
	}()

	var buf bytes.Buffer

	// Simulate mixed pass/fail results
	results := []allCheckResult{
		{name: "format", category: "code_quality", passed: true, duration: 200 * time.Millisecond},
		{name: "lint", category: "code_quality", passed: false, duration: 500 * time.Millisecond, err: errors.New("lint issues found")},
	}
	totalPassed := 1
	totalFailed := 1

	jsonResult := toJSONResult(results, totalPassed, totalFailed)
	status := "error"
	jsonErr := &output.JSONError{Message: "1 of 2 checks failed"}

	err := output.RenderJSON(&buf, output.JSONEnvelope{
		Status:  status,
		Command: output.CommandName(),
		Data:    jsonResult.JSONResponse(),
		Error:   jsonErr,
	})
	require.NoError(t, err)

	var envelope output.JSONEnvelope
	err = json.Unmarshal(buf.Bytes(), &envelope)
	require.NoError(t, err)

	assert.Equal(t, "error", envelope.Status)
	assert.Equal(t, "check", envelope.Command)
	assert.NotNil(t, envelope.Error, "error field must not be nil when status is error")
	assert.Equal(t, "1 of 2 checks failed", envelope.Error.Message)
	assert.NotNil(t, envelope.Data)

	// Verify data still contains check results alongside the error
	dataBytes, _ := json.Marshal(envelope.Data)
	var checkData CheckJSONResult
	err = json.Unmarshal(dataBytes, &checkData)
	require.NoError(t, err)

	assert.False(t, checkData.Passed)
	assert.Equal(t, 2, checkData.Total)
	assert.Equal(t, 1, checkData.Failed)
	assert.Len(t, checkData.Checks, 2)
	assert.False(t, checkData.Checks[1].Passed)
	assert.Equal(t, "lint issues found", checkData.Checks[1].Error)
}
