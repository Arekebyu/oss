FROM python:3.11-slim
WORKDIR /app
COPY python_ml/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY python_ml/ .
EXPOSE 50051
CMD ["python", "server.py"]