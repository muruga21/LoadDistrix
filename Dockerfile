FROM golang:1.22.5-alpine
WORKDIR /app
COPY . .
RUN go build -o loaddistrix
CMD ["./loaddistrix"]
