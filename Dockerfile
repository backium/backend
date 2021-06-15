FROM golang:1.16.5-alpine

# Set some configurations
ENV BACKIUM_DB_URI  "missing db uri" 
ENV BACKIUM_DB_NAME  "missing db name" 
ENV BACKIUM_APP_PORT  "missing app port" 
ENV BACKIUM_REDIS_URI  "missing redis uri" 
ENV BACKIUM_REDIS_PASSWORD  "missing redis password"

# Make the source code path
RUN mkdir -p /app

# Add all source code
ADD . /app

WORKDIR /app

# Run the Go installer
RUN go install

# Indicate the binary as our entrypoint
ENTRYPOINT /go/bin/backend

# Expose your port
EXPOSE 8080
