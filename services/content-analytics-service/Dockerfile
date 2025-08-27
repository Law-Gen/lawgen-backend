# --- Build stage ---
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/server

# --- Runtime stage ---
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/app
COPY configs/config.example.yaml /app/configs/config.yaml
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]