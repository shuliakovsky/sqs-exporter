FROM golang:latest

RUN useradd -m -s /bin/bash appuser
WORKDIR /app
COPY . .
RUN chown -R appuser /app
USER appuser
RUN go mod download
EXPOSE 8080
ENTRYPOINT ["go", "run", "main.go", "config.go", "metrics.go", "sqs.go", "-config", "./config.yaml"]
