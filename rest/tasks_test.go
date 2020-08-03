package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bitbucket.org/kleinnic74/photos/tasks"
	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
)

func TestGetTaskDefinitions(t *testing.T) {
	executor := tasks.NewDummyTaskExecutor()
	api := NewTaskHandler(executor)
	router := mux.NewRouter()
	api.InitRoutes(router)

	req, _ := http.NewRequest("GET", "/taskdefinitions", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	response := rr.Result()
	checkResponseCode(t, http.StatusOK, response)
	assertContentType(t, "application/json", response)
	var result struct {
		Data []tasks.TaskDefinition
	}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %s", err)
	}
}

func TestPostTask(t *testing.T) {
	executor := tasks.NewDummyTaskExecutor()
	api := NewTaskHandler(executor)
	router := mux.NewRouter()
	api.InitRoutes(router)

	payload := `{"type":"import","parameters":{"importdir":"from","dryrun":true}}`
	req, _ := http.NewRequest(http.MethodPost, "/tasks", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	response := rr.Result()
	t.Logf("Response body:[[%s]]", rr.Body)
	checkResponseCode(t, http.StatusAccepted, response)
	assertContentType(t, "application/json", response)
	var result tasks.Execution
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %s", err)
	}
	t.Logf("Answer=%v", result)
}

func TestTaskJSON(t *testing.T) {
	data := []struct {
		json string
		task tasks.Task
	}{
		{json: `{"type":"import","parameters":{"dryrun":true,"importdir":"/path/to/dir"}}`,
			task: tasks.NewImportTaskWithParams(true, "/path/to/dir")},
		{json: `{"type":"import","parameters":{"importdir":"/path/to/dir"}}`,
			task: tasks.NewImportTaskWithParams(false, "/path/to/dir")},
	}
	for _, d := range data {
		task, err := parseTask(strings.NewReader(d.json))
		if err != nil {
			t.Fatalf("Failed to parse '%s': %s", d.json, err)
		}
		assert.Equal(t, d.task, task, "Parsed task does not match expectation")
	}
}

func assertContentType(t *testing.T, expected string, response *http.Response) {
	actual := response.Header.Get("Content-Type")
	if actual != expected {
		t.Errorf("Bad Content-Type: expected '%s', got '%s'", expected, actual)
	}
}
