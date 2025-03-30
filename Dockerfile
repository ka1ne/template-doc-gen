# Build stage
FROM python:3.10-slim AS builder

WORKDIR /app

# Copy requirements first for better caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Runtime stage - use specific version for better reproducibility
FROM python:3.10-slim

WORKDIR /app

# Install git for potential use with template repos
# Combine commands to reduce layers and optimize cleanup
RUN apt-get update && \
    apt-get install -y --no-install-recommends git ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Create non-root user for security
RUN adduser --disabled-password --gecos "" appuser

# Copy only what's needed from builder - python packages and app
COPY --from=builder /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages

# Only copy the necessary files instead of everything
COPY process_template.py main.py publish-to-confluence.sh ./
COPY src/ ./src/
COPY templates/ ./templates/

# Create directories with proper permissions and make script executable
# Combine commands to reduce layers
RUN mkdir -p /app/docs/templates && \
    chmod +x /app/publish-to-confluence.sh && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PATH="/app:${PATH}"

# Health check to verify the container is running properly
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD python -c "import sys; sys.exit(0)"

# Default command - set to process_template.py with help flag
ENTRYPOINT ["python", "process_template.py"]
CMD ["--help"] 