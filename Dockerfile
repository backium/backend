FROM golang:1.16.5-alpine

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

# Indicate the binary as our entrypoint
ENTRYPOINT /go/bin/backend

# Expose your port
EXPOSE 8080
