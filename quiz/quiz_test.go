package quiz

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartQuizHandlerWithoutUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/quiz/start", nil)
	rr := httptest.NewRecorder()

	StartQuizHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestStartQuizHandlerWithUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/quiz/start?user=tester", nil)
	rr := httptest.NewRecorder()

	StartQuizHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestAnswerQuizHandlerCorrect(t *testing.T) {
	// Reset sesi agar independen
	ResetSession("tester")

	// Simulasikan permintaan untuk memulai quiz (set CurrentIndex ke 0)
	StartQuizHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/quiz/start?user=tester", nil))

	answer := AnswerRequest{
		UserID: "tester",
		Answer: "Paris",
	}
	body, _ := json.Marshal(answer)
	req := httptest.NewRequest(http.MethodPost, "/quiz/answer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	AnswerQuizHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp AnswerResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.Correct {
		t.Error("expected answer to be correct")
	}
}

func TestResetHandler(t *testing.T) {
	// Buat sesi dummy
	GetOrCreateSession("john")

	req := httptest.NewRequest(http.MethodGet, "/quiz/reset?user=john", nil)
	rr := httptest.NewRecorder()

	ResetHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}
