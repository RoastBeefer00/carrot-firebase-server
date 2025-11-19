######################################
# STAGE 1: Frontend asset generation
######################################
FROM debian:12-slim AS frontend-builder
WORKDIR /app

# Install curl for downloading Tailwind CLI
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Download standalone Tailwind CLI
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && \
        chmod +x tailwindcss-linux-x64

# Copy Tailwind config and CSS files
COPY dist/main.css ./dist/
COPY views/ ./views/

# Generate CSS with standalone Tailwind CLI
RUN ./tailwindcss-linux-x64 -i ./dist/main.css -o ./dist/tailwind.css --minify

######################################
# STAGE 2: Templ generation and Go build
######################################
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install Templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy the generated Tailwind CSS from the frontend stage
COPY --from=frontend-builder /app/dist/tailwind.css ./dist/

# Generate Templ templates
RUN templ generate

# Build the Go application
RUN CGO_ENABLED=0 go build -o app .

######################################
# STAGE 3: Final runtime image
######################################
FROM alpine:3.18
WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /app/app .

# Copy static assets
COPY --from=builder /app/dist ./dist

# Expose the port your application runs on
EXPOSE 8080

# Run the application
CMD ["./app"]
