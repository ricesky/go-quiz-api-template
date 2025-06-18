# Studi Kasus: Quiz API

## Capaian Pembelajaran

* Peserta dapat membangun web service sederhana menggunakan `net/http` di Go.
* Peserta dapat mengimplementasikan REST API menggunakan struct, slice, dan fungsi handler.
* Peserta dapat mengembangkan manajemen sesi sederhana menggunakan `map` dan `struct`.

---

## Deskripsi

Quiz API adalah layanan web services sederhana yang menyediakan:

* Daftar soal quiz (`/quiz/start`)
* Pemeriksaan jawaban (`/quiz/answer`)
* Manajemen progres user dengan sesi (`/quiz/reset`, `/quiz/start?user`, `/quiz/answer`)

---

## Prasyarat

* Go 1.24
* VSCode with Go extension
* Delve debugger
* REST client seperti Postman atau `curl`

---

## Struktur Direktori

```
go-quiz-api/
├── go.mod
├── main.go
└── quiz/
    ├── quiz.go
    └── session.go
```

---

## Langkah 1: Membuat Project

Masuk ke dalam repositori dan jalankan perintah berikut untuk melakukan inisialisasi modul.

```bash
go mod init go-quiz-api
mkdir quiz
touch main.go
touch quiz/quiz.go
```

---

## Langkah 2: Membuat Basic HTTP Server

### `main.go`

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is up"))
	})

	fmt.Println("Server started at :8787")
	log.Fatal(http.ListenAndServe(":8787", nil))
}
```

**Uji coba:**

```bash
curl http://localhost:8787/healthcheck
```

---

## ▶️ Langkah 3: Membuat Endpoint Start Quiz (Tanpa Sesi)

### Tujuan:

Mengirimkan seluruh daftar soal quiz kepada client.

### `quiz/quiz.go`

```go
package quiz

import (
	"encoding/json"
	"net/http"
)

type Question struct {
	ID       int
	Question string
	Answer   string
}

var questionBank = []Question{
	{ID: 1, Question: "What is the capital of France?", Answer: "Paris"},
	{ID: 2, Question: "2 + 2 = ?", Answer: "4"},
	{ID: 3, Question: "Go is statically typed. (true/false)", Answer: "true"},
}

func StartQuizHandler(w http.ResponseWriter, r *http.Request) {
	type QuestionItem struct {
		ID       int    `json:"id"`
		Question string `json:"question"`
	}

	var qlist []QuestionItem
	for _, q := range questionBank {
		qlist = append(qlist, QuestionItem{ID: q.ID, Question: q.Question})
	}

	json.NewEncoder(w).Encode(qlist)
}
```

### Tambahkan ke `main.go`

```go
import "go-quiz-api/quiz"

func main() {
	// kode sebelumnya ...

	http.HandleFunc("/quiz/start", quiz.StartQuizHandler)

	fmt.Println("Server started at :8787")
	log.Fatal(http.ListenAndServe(":8787", nil))
}

```

**Uji coba:**

```bash
curl http://localhost:8787/quiz/start
```

---

## Langkah 4: Membuat Endpoint Answer Quiz (Tanpa Sesi)

### Tujuan:

Memeriksa jawaban yang dikirim user berdasarkan `ID` soal.

### Tambahkan di `quiz/quiz.go`

```go
import (
	"strings"
)

// kode sebelumnya ...

type AnswerRequest struct {
	ID     int    `json:"id"`
	Answer string `json:"answer"`
}

type AnswerResponse struct {
	Correct bool   `json:"correct"`
	Message string `json:"message"`
}

func AnswerQuizHandler(w http.ResponseWriter, r *http.Request) {
	var ans AnswerRequest
	err := json.NewDecoder(r.Body).Decode(&ans)
	if err != nil || ans.ID <= 0 || ans.ID > len(questionBank) {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	correctAnswer := strings.ToLower(questionBank[ans.ID-1].Answer)
	givenAnswer := strings.ToLower(strings.TrimSpace(ans.Answer))

	var response AnswerResponse
	if correctAnswer == givenAnswer {
		response = AnswerResponse{Correct: true, Message: "Correct!"}
	} else {
		response = AnswerResponse{Correct: false, Message: "Wrong answer"}
	}

	json.NewEncoder(w).Encode(response)
}
```

### Tambahkan ke `main.go`

```go
import "go-quiz-api/quiz"

func main() {
	// kode sebelumnya ...

	http.HandleFunc("/quiz/answer", quiz.AnswerQuizHandler)

	fmt.Println("Server started at :8787")
	log.Fatal(http.ListenAndServe(":8787", nil))
}
```

**Uji coba:**

```bash
curl -X POST http://localhost:8787/quiz/answer \
  -H "Content-Type: application/json" \
  -d '{"id": 1, "answer": "Paris"}'
```

---

## Langkah 5: Menambahkan Manajemen Sesi

### Tujuan:

Memungkinkan tiap user memiliki progres quiz masing-masing.

### Buat `quiz/session.go`

```go
package quiz

type Session struct {
	UserID        string
	CurrentIndex  int
	Score         int
	Finished      bool
}

var sessions = make(map[string]*Session)

func GetOrCreateSession(userID string) *Session {
	if s, ok := sessions[userID]; ok {
		return s
	}
	s := &Session{UserID: userID}
	sessions[userID] = s
	return s
}

func ResetSession(userID string) {
	delete(sessions, userID)
}
```

---

## Langkah 6: Modifikasi Endpoint Start Quiz Agar Menggunakan Sesi

```go
func StartQuizHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	session := GetOrCreateSession(userID)
	if session.Finished {
		http.Error(w, "Quiz already finished", http.StatusForbidden)
		return
	}

	q := questionBank[session.CurrentIndex]
	json.NewEncoder(w).Encode(q)
}
```

**Uji coba:**

```bash
curl http://localhost:8787/quiz/start?user=john
```

---

## Langkah 7: Modifikasi Endpoint Answer Quiz Agar Gunakan Sesi

```go
type Question struct {
	ID       int
	Question string
	Answer   string `json:"-"`
}

type AnswerRequest struct {
	UserID string `json:"user_id"`
	Answer string `json:"answer"`
}

type AnswerResponse struct {
	Correct       bool      `json:"correct"`
	Message       string    `json:"message"`
	Score         int       `json:"score"`
	QuestionsLeft int       `json:"questions_left"`
	Finished      bool      `json:"finished"`
	NextQuestion  *Question `json:"next_question,omitempty"`
}

func AnswerQuizHandler(w http.ResponseWriter, r *http.Request) {
	var ans AnswerRequest
	err := json.NewDecoder(r.Body).Decode(&ans)
	if err != nil || ans.UserID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session := GetOrCreateSession(ans.UserID)
	if session.Finished {
		http.Error(w, "Quiz already finished", http.StatusForbidden)
		return
	}

	correctAns := strings.ToLower(questionBank[session.CurrentIndex].Answer)
	givenAns := strings.ToLower(strings.TrimSpace(ans.Answer))

	var resp AnswerResponse
	if correctAns == givenAns {
		session.Score++
		resp.Correct = true
		resp.Message = "Correct!"
	} else {
		resp.Correct = false
		resp.Message = "Incorrect."
	}

	session.CurrentIndex++
	if session.CurrentIndex >= len(questionBank) {
		session.Finished = true
		resp.Finished = true
		resp.QuestionsLeft = 0
	} else {
		resp.NextQuestion = &questionBank[session.CurrentIndex]
		resp.QuestionsLeft = len(questionBank) - session.CurrentIndex
	}
	resp.Score = session.Score

	json.NewEncoder(w).Encode(resp)
}
```

**Uji coba:**

```bash
curl -X POST http://localhost:8787/quiz/answer \
  -H "Content-Type: application/json" \
  -d '{"user_id": "john", "answer": "Paris"}'
```

---

## Langkah 8: Tambahkan Endpoint Reset Sesi

```go
func ResetHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	ResetSession(userID)
	w.Write([]byte("Session reset"))
}
```

**Tambahkan ke `main.go`:**

```go
http.HandleFunc("/quiz/reset", quiz.ResetHandler)
```

**Uji coba:**

```bash
curl http://localhost:8787/quiz/reset?user=john
```
---

**Selesai**
