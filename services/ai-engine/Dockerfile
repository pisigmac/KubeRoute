FROM python:3.11-slim

WORKDIR /app
RUN pip install uv
COPY pyproject.toml .
RUN uv pip install --system -e .
COPY ai_engine/ ./ai_engine/
COPY config/ ./config/

EXPOSE 8081
CMD ["uvicorn", "ai_engine.main:app", "--host", "0.0.0.0", "--port", "8081"]
