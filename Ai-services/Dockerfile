FROM python:3.12-slim

# Set the working directory in the container to the "app" folder
WORKDIR /app

# Copy the requirements file and install dependencies first for caching
# This is an optimization so subsequent builds don't re-download packages
COPY app/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ .

# Expose the port that FastAPI will run on
EXPOSE 9000

# Run the application using uvicorn from the app directory, cause i import some of the azure util and engine functions as a modelu
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]