######################################
# STAGE 1: Frontend asset generation
######################################
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package.json and package-lock.json for npm install
COPY package*.json ./

# Copy Tailwind and CSS related files
COPY tailwind.config.js ./
COPY dist/main.css ./dist/
COPY views/ ./views/

# Install npm dependencies
RUN npm install

# Run Tailwind to generate CSS
RUN npx tailwindcss -i ./dist/main.css -o ./dist/tailwind.css

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

# Copy static assets and templates
COPY --from=builder /app/dist ./dist
# COPY --from=builder /app/templates ./templates
# COPY --from=builder /app/static ./static

# Expose the port your application runs on
EXPOSE 8080

# Run the application
CMD ["./app"]
