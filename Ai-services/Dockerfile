# Use Python slim image
FROM python:3.12-slim

# Set working directory
WORKDIR /app

# Copy requirements from root of Ai-services
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy app folder into container
COPY ./app ./app

# Expose the port (optional)
EXPOSE 8000

# Run Uvicorn using Azure $PORT
CMD ["sh", "-c", "uvicorn app.main:app --host 0.0.0.0 --port ${PORT:-8000}"]
