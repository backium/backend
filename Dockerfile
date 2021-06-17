FROM golang:1.16.5-alpine as builder

# Set some configurations
ENV BACKIUM_DB_URI  "mongodb://localhost:27017" 
ENV BACKIUM_DB_NAME  "testing" 
ENV BACKIUM_APP_PORT  "8080" 
ENV BACKIUM_REDIS_URI  "localhost:6379" 
ENV BACKIUM_REDIS_PASSWORD  ""

# Make the source code path
RUN mkdir -p /app

# Add all source code
ADD . /app

WORKDIR /app

# Run the Go installer
RUN go install

# FROM scratch
FROM alpine:latest

RUN mkdir -p /app

WORKDIR /app

# Copy bynary from builder
COPY --from=builder /go/bin/backend /app/

# Copy configuration files from builder
COPY --from=builder /app/app.env /app/

# Expose your port
EXPOSE 8080

# Indicate the binary as our entrypoint
ENTRYPOINT [ "/app/backend" ]