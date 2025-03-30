FROM python:3.10-slim

LABEL maintainer="Kaine@enterpriseautomation.co.uk"
LABEL version="0.0.3-alpha"


# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# Create app directory
WORKDIR /app

# Install dependencies first (for better layer caching)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application files
COPY . .

# Set up default configuration
ENV SOURCE_DIR=/app/templates \
    OUTPUT_DIR=/app/docs/output \
    FORMAT=html \
    VERBOSE=false \
    VALIDATE_ONLY=false

# Create output directory
RUN mkdir -p /app/docs/output && chmod 777 /app/docs/output

# Define entrypoint and default command
ENTRYPOINT ["python", "process_template.py"]
CMD ["--source", "/app/templates", "--output", "/app/docs/output", "--format", "html"] 