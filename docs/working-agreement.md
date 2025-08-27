# Working Agreement (Phase 1: no RabbitMQ)

- Monorepo: services live under `services/`.
- Single service for now: `content-analytics-service`.
- API Gateway handles authN/authZ and forwards identity via headers:
  - `X-User-Id`, `X-User-Roles` (comma-separated).
- Only Admin uploads files to cloud (Azure Blob planned). This service stores metadata and enables browsing/search.
- Ownership split:
  - You: Content + Feedback (implement).
  - Teammate: Legal Entity + Analytics (scaffolded, to implement).
- No RabbitMQ for now; event consumers can be added later without breaking public APIs.
- Error schema:
  ```
  { "code": "ERROR_CODE", "message": "human-readable", "details": { ... } }
  ```
