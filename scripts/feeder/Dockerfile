FROM golang:1.16.5-alpine

# Make the source code path
RUN mkdir -p /app

# Add all source code
ADD . /app

WORKDIR /app

RUN apk add git

RUN git config --global url."https://echovl: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@github.com/".insteadOf "https://github.com/"

RUN go env -w GOPRIVATE=github.com/backium

RUN echo "127.0.0.1\tmongodb" | (tee -a /etc/hosts) > /dev/null

RUN go mod init feeder &&  go get "github.com/backium/backend@10e869ae6cfc36aadcfb1df9b7a869cf49433187" && go mod tidy

RUN go build -o main .

ENTRYPOINT ["sh", "-c", "./main --mongo-uri $DB_URI --mongo-name $DB_NAME"]
