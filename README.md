# Backend-Dashboard

## Requirements

Docker, docker compose and golang.

## How to start

To start the server in development mode run the following command.

```sh
$ docker-compose up
```

To add mock data, first add mongodb host to `/etc/hosts`

```sh
$ echo "127.0.0.1\tmongodb" | (sudo tee -a /etc/hosts) > /dev/null
```

Finally, run the script to generate mock data.

```sh
$ go run ./scripts/setup.go
```

