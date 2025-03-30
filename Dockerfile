# Build stage
FROM python:3.10-slim AS builder

WORKDIR /app

# Copy requirements first for better caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Runtime stage
FROM python:3.10-slim

WORKDIR /app

# Install git for potential use with template repos
RUN apt-get update && \
    apt-get install -y --no-install-recommends git && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN adduser --disabled-password --gecos "" appuser

# Copy only what's needed from builder
COPY --from=builder /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages
COPY --from=builder /app /app

# Create directories with proper permissions and make script executable
RUN mkdir -p /app/docs/templates && \
    chmod +x /app/publish-to-confluence.sh && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PATH="/app:${PATH}"

# Default command - set to process_template.py with help flag
ENTRYPOINT ["python", "process_template.py"]
CMD ["--help"] 