FROM python:3.12-slim

# Set the working directory in the container to the "app" folder
WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy all the application code into the container
COPY app/ .

# Expose the port that FastAPI will run on
EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]