## Installation
1. **Clone the repository:**
   ```bash
   git clone <repo-url>
   cd lawgen-backend/chat-service
   ```
2. **Configure environment:**
   - Copy `.env.example` to `.env` and set your environment variables.
3. **Install dependencies:**
   ```bash
   go mod tidy
   ```

## Running the Service
```bash
go run .
```
The service will start on the port specified in your `.env` file.


# LawGen Chat Service

This service provides RESTful APIs for quizzes, chat, and voice chat, designed for integration with LawGen frontend apps and other backend services.

## Features
- Quiz and category CRUD operations
- Add/update/delete questions
- Submit quiz answers and get scores
- Pagination for categories and quizzes
- Chat with legal assistant (SSE streaming, session history, sources)
- **Voice chat with legal assistant (audio in, audio out, Amharic/English supported)**

# Performance

We benchmarked the LawGen backend using [k6](https://k6.io/) before and after key optimizations (MongoDB indexes, gzip compression).

**Test Setup:**
- Tool: k6
- Environment: Localhost, 10 virtual users (VUs), 10 seconds per test
- Endpoints tested: `/api/v1/chats/query`, `/api/v1/quizzes/categories`
- Logs: See `perf_logs/original_*` and `perf_logs/opt_*` for full details

### Results

| Endpoint                  | Test        | Avg Latency | 95th Percentile | Max Latency | RPS   | Success Rate |
|---------------------------|-------------|-------------|-----------------|-------------|-------|--------------|
| /api/v1/chats/query       | Original    | 792.03ms    | 1.29s           | 2.04s       | 5.34  | 100%         |
| /api/v1/chats/query       | Optimized   | 771.85ms    | 1.52s           | 1.53s       | 5.28  | 100%         |
| /api/v1/quizzes/categories| Original    | 9.56ms      | 21.24ms         | 38.73ms     | 9.88  | 100%         |
| /api/v1/quizzes/categories| Optimized   | 4.53ms      | 7.96ms          | 9.60ms      | 9.94  | 100%         |

#### Key Improvements

- **Quiz endpoints:** Latency cut by more than half (avg 9.56ms → 4.53ms), and max latency dropped from 38.73ms to 9.60ms.
- **Chat endpoint:** Slight improvement in average and max latency, and more consistent high-percentile response times.
- **System scalability:** Backend maintains 100% success rate and stable throughput under load.

#### Example Log Snippet

```
Original chat:
http_req_duration: avg=792.03ms min=310.1ms med=680.05ms max=2.04s p(90)=1.11s p(95)=1.29s

Optimized chat:
http_req_duration: avg=771.85ms min=372.27ms med=625.55ms max=1.53s p(90)=1.51s p(95)=1.52s

Original quiz:
http_req_duration: avg=9.56ms min=2.52ms med=8.34ms max=38.73ms p(90)=19.68ms p(95)=21.24ms

Optimized quiz:
http_req_duration: avg=4.53ms min=1.06ms med=4.67ms max=9.6ms p(90)=6.99ms p(95)=7.96ms
```

> You can update this table as you further optimize or scale your system. For more details, see the full logs in the `perf_logs/` directory.
## Testing
- For quizzes: Import `quiz_api_collection.json` into Postman for ready-to-use API tests.
- For chat: Open `index.html` in your browser at `http://localhost:8080/` after running the service. This provides a production-ready chat UI that interacts with the backend via SSE and REST endpoints.
- For **voice chat**: Use the "Voice Chat" panel in `index.html` or send a `multipart/form-data` POST to `/api/v1/chats/voice-query` (see below).

---

## API Endpoints

### Categories, Quizzes, Questions, Quiz Submission
*(See previous sections for details.)*

---

### Chat Service (Text)
- `POST /api/v1/chats/query`: Send a chat message and receive streamed response (SSE)
  - Request: `{ "sessionId": "<optional>", "query": "<message>", "language": "<optional>" }`
  - Response: SSE stream of `{ text, sources, is_complete, suggested_questions }`
- `GET /api/v1/chats/sessions`: List chat sessions for authenticated user
- `GET /api/v1/chats/sessions/:sessionId/messages`: Get messages for a session

#### Usage Example (Chat Query)
```bash
curl -X POST http://localhost:8080/api/v1/chats/query \
  -H "Content-Type: application/json" \
  -d '{"sessionId": "<optional>", "query": "What is contract law?", "language": "en"}'
```

---

### Voice Chat Service

- `POST /api/v1/chats/voice-query`
  - **Request:** `multipart/form-data` with fields:
    - `file`: audio file (WAV/MP3, Amharic or English speech)
    - `language`: `"en"` or `"am"`
    - (optional) `sessionId`, `userId`, `planId`
  - **Response:** `audio/mpeg` (MP3 audio, same language as request)
  - **No JSON is returned.** The response is a raw audio file.

#### Example (using curl)
```bash
curl -X POST http://localhost:8080/api/v1/chats/voice-query \
  -F "file=@sample.wav" \
  -F "language=am" \
  --output response.mp3
```

