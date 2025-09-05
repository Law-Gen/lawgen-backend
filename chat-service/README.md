go run .

# Chat Service

This service provides RESTful APIs for quizzes, chat, and voice chat, designed for integration with LawGen frontend apps and other backend services.

## Features
- Quiz and category CRUD operations
- Add/update/delete questions
- Submit quiz answers and get scores
- Pagination for categories and quizzes
- Chat with legal assistant (SSE streaming, session history, sources)
- **Voice chat with legal assistant (audio in, audio out, Amharic/English supported)**

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

#### Frontend Integration
- Use `fetch` with `FormData` and expect a raw audio response.
- Example:
  ```js
  const resp = await fetch('/api/v1/chats/voice-query', { method: 'POST', body: formData });
  const audioBlob = await resp.blob();
  voicePlayer.src = URL.createObjectURL(audioBlob);
  voicePlayer.play();
  ```
- **Do not parse the response as JSON.** The backend returns only audio.

#### Supported Languages
- `"en"`: English
- `"am"`: Amharic

#### Error Handling
- On error, the backend returns a JSON error object with an appropriate HTTP status code.
- On success, the response is always `audio/mpeg`.

---

## License
MIT
