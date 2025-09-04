# Chat Service

This service provides RESTful APIs for managing quizzes and categories, designed for integration with other backend services.

## Features
- Quiz and category CRUD operations
- Add/update/delete questions
- Submit quiz answers and get scores
- Pagination for categories and quizzes
- Chat with legal assistant (SSE streaming, session history, sources)

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
  - You can send messages, view session history, and see sources used in answers.

## API Endpoints
### Categories
- `GET /categories`: List categories (pagination: `page`, `limit`)
- `POST /categories`: Create category `{ name }`
- `PUT /categories/:categoryId`: Update category `{ name }`
- `DELETE /categories/:categoryId`: Delete category

### Quizzes
- `GET /categories/:categoryId/quizzes`: List quizzes by category (pagination)
- `POST /quizzes`: Create quiz `{ category_id, name, description }`
- `PUT /quizzes/:quizId`: Update quiz `{ name, description }`
- `DELETE /quizzes/:quizId`: Delete quiz
- `GET /quizzes/:quizId`: Get quiz details

### Questions
- `GET /quizzes/:quizId/questions`: List questions for a quiz
- `POST /quizzes/:quizId/questions`: Add question `{ text, options, correct_option }`
- `PUT /quizzes/:quizId/questions/:questionId`: Update question `{ text, options, correct_option }`
- `DELETE /quizzes/:quizId/questions/:questionId`: Delete question

### Quiz Submission
- `POST /quizzes/:quizId/submit`: Submit answers `{ "questionId": "selectedOption", ... }`
  - Returns: `{ score, total_question }`

### Chat Service
- `POST /api/v1/chats/query`: Send a chat message and receive streamed response (SSE)
  - Request: `{ "sessionId": "<optional>", "query": "<message>", "language": "<optional>" }`
  - Response: SSE stream of `{ text, sources, is_complete, suggested_questions }`
- `POST /api/chats`: Create a new chat session
  - Request: `{ "user_id": "<userId>", "topic": "<topic>" }`
  - Response: `{ session_id }`
- `GET /api/v1/chats/sessions`: List chat sessions for authenticated user
  - Query params: `page`, `limit`
  - Response: `{ sessions: [...], total, page, limit }`
- `GET /api/v1/chats/sessions/:sessionId/messages`: Get messages for a session
  - Response: `{ messages: [...] }`

#### Usage Example (Chat Query)
```bash
curl -X POST http://localhost:8080/api/v1/chats/query \
  -H "Content-Type: application/json" \
  -d '{"sessionId": "<optional>", "query": "What is contract law?", "language": "en"}'
```

#### SSE Streaming
- The `/api/v1/chats/query` endpoint streams responses word-by-word using Server-Sent Events (SSE).
- Parse each SSE event for `{ text }` and display incrementally.
- Final event includes `is_complete: true` and `sources` used in the answer.

## License
MIT
