# Chat Service

This service provides RESTful APIs for managing quizzes and categories, designed for integration with other backend services.

## Features
- Quiz and category CRUD operations
- Add/update/delete questions
- Submit quiz answers and get scores
- Pagination for categories and quizzes

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

## Testing
- Import `quiz_api_collection.json` into Postman for ready-to-use API tests.

## License
MIT
