FROM python:3.12-slim

# Set the working directory in the container to /app
WORKDIR /app

# Copy and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy all Python files into the container
COPY . .

# Expose the port that FastAPI will run on
EXPOSE 8000

# Use main.py at the same level as WORKDIR
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
