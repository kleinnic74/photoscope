package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"bitbucket.org/kleinnic74/photos/rest/cursor"
	"bitbucket.org/kleinnic74/photos/tasks"
	"github.com/gorilla/mux"
)

type TaskHandler struct {
	executor tasks.TaskExecutor
}

func NewTaskHandler(executor tasks.TaskExecutor) *TaskHandler {
	return &TaskHandler{executor: executor}
}

func (h *TaskHandler) InitRoutes(r *mux.Router) {
	r.HandleFunc("/taskdefinitions", h.getTaskDefinitions).Methods("GET")
	r.HandleFunc("/tasks", h.postTask).Methods("POST")
	r.HandleFunc("/tasks", h.listTasks).Methods("GET")
}

func (h *TaskHandler) getTaskDefinitions(w http.ResponseWriter, r *http.Request) {
	defined := tasks.DefinedTasks()
	respondWithJSON(w, http.StatusOK, &simplePayload{Data: defined})
}

type task struct {
	Type       string          `json:"type"`
	Parameters json.RawMessage `json:"parameters"`
}

func (h *TaskHandler) postTask(w http.ResponseWriter, r *http.Request) {
	task, err := parseTask(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	execution, err := h.executor.Submit(r.Context(), task)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, err)
	}
	respondWithJSON(w, http.StatusAccepted, execution)
}

func (h *TaskHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.executor.ListTasks(r.Context())
	respondWithJSON(w, http.StatusOK, cursor.Unpaged(tasks))
}

func parseTask(in io.Reader) (t tasks.Task, err error) {
	var tmp task
	err = json.NewDecoder(in).Decode(&tmp)
	if err != nil {
		return
	}
	t, err = tasks.CreateTask(tmp.Type)
	if err != nil {
		return
	}
	json.Unmarshal(tmp.Parameters, t)
	return
}
