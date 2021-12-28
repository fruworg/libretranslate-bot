FROM golang:last

WORKDIR /app
COPY . .

CMD ["go run ."]
