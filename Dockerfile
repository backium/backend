FROM golang:1.16.5-alpine

# Set some configurations
ENV BACKIUM_DB_URI  "mongodb+srv://dev:dev@uibox-dev.wtl7h.mongodb.net/dev?retryWrites=true&w=majority" 
ENV BACKIUM_DB_NAME  "testing" 
ENV BACKIUM_APP_PORT  8080 
ENV BACKIUM_REDIS_URI  "redis-19153.c244.us-east-1-2.ec2.cloud.redislabs.com:19153" 
ENV BACKIUM_REDIS_PASSWORD  "pAHrQLGobMNzfoLeIPv3vdqYLCNZHS79"

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
