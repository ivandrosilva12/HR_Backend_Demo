FROM golang:1.24.1 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o hrms ./main.go

# Final image
FROM debian:bookworm-slim
WORKDIR /root/
COPY --from=builder /app/hrms .

# (Opcional) Se não usa filesystem local, pode remover esta linha:
# COPY --from=builder /app/uploads ./uploads

EXPOSE 8080
CMD ["./hrms"]