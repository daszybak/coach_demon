Of course! Here's a clean **README.md** for your project, and afterward, a **summary you can copy for your CV**.

---

# coach_demon

**Competitive Programming AI Coach**  
A backend service that monitors your coding during problem solving and sends feedback powered by OpenAI.  
Includes real-time WebSocket communication, MongoDB storage, problem fetching from Codeforces, and smart coaching summaries.

---

## ✨ Features

- **WebSocket Server** — real-time editor feedback loop
- **MongoDB Storage** — snapshots of code, thoughts, feedbacks, proofs
- **Problem Fetcher** — scrapes Codeforces problem statements automatically
- **AI Feedback Engine** — powered by OpenAI structured responses
- **Journey and Integration Tests** — full flow automated test suites
- **Dockerized Setup** — including MongoDB, Fetcher, and Test runner
- **CI-like Test Execution** — runs tests in isolated containers with full volume binding for reports
- **Extensible Architecture** — modular fetcher, AI, and storage components
- **Clean Makefile Commands** — fast local builds and test workflows

---

## 🐳 Docker Compose Setup

Spin up everything:

```bash
docker compose up --build
```

Components:
- `mongo` — MongoDB database
- `coach_demon` — main backend application
- `fetcher` — Browserless Chrome container for problem fetching
- `tests` — containerized test runner (journey + integration tests)

---

## ⚡ Running Tests

First, **build the test runner**:

```bash
make build
```

Then run all tests:

```bash
make test-all
```

- Integration tests live under `tests/integration`
- Journey (end-to-end) tests live under `tests/journey`
- Reports are automatically generated into `tests/reports/`

You can also run only specific test suites:

```bash
make test-integration
make test-journey
```

---

## 📁 Project Structure

```plaintext
cmd/coach_demon/        → main entrypoint
internal/app/           → runtime dependency injection
internal/fetcher/       → Codeforces problem fetcher
internal/openai/        → OpenAI feedback client
internal/storage/       → MongoDB management
internal/server/        → HTTP and WebSocket handlers
tests/integration/      → integration (live) tests
tests/journey/          → journey (E2E) tests
Dockerfile              → multi-stage build (runtime & tests)
docker-compose.yml      → runtime and test orchestration
Makefile                → simple CI automation
```

---

## 🧹 How to Add a New Test

Inside the `tests/integration/` or `tests/journey/` folder:

```bash
//go:build integration

package integration

import (
    "testing"
)

func TestMyNewCase(t *testing.T) {
    t.Log("Hello new test!")
}
```

Then:

```bash
make test-integration
# or
make test-journey
```

✅ Report will be saved into `tests/reports/`.

---

## 📜 Summary for CV

**Backend Developer (Go, Docker, MongoDB)**  
Built a real-time AI coaching platform for competitive programmers, combining WebSocket-based feedback loops, OpenAI integration, Codeforces problem fetching, and MongoDB persistence. Designed a fully containerized system with Docker Compose and automated journey + integration testing inside Dockerized environments. Developed a clean Makefile pipeline for building, running, and testing the system, ensuring CI/CD-like workflows locally.

---

Would you also like a slightly shorter version (like a one-liner) you can use on your CV if you want it even tighter? 🚀  
(Just say “yes, short version!”)