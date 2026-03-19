# Stage 1: Build CSS with Tailwind
FROM node:20-alpine AS css-builder
WORKDIR /app
COPY package*.json ./ 2>/dev/null || true
RUN npm install -g tailwindcss

COPY . .
RUN tailwindcss -i ./web/static/css/input.css -o ./web/static/css/app.css --minify

# Stage 2: Build the Go application
FROM golang:1.24-alpine AS go-builder
ENV TZ=America/Sao_Paulo
RUN apk add --no-cache upx make tzdata
WORKDIR /app

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Copy the generated CSS from the previous stage
COPY --from=css-builder /app/web/static/css/app.css ./web/static/css/app.css

# Build the binary
RUN make build-docker-prod

# Stage 3: Final minimal image
FROM alpine:3.21
ENV TZ=America/Sao_Paulo
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=go-builder /app/bin/gorundeck .
COPY config.toml.example ./config.toml.example
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expose the default port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]

# Default command starting the server
CMD ["./gorundeck", "serve"]
