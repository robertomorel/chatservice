# Chatservice

This project is to implement the backend for an application to integrate Whatsapp with the OpenAI API (ChatGPT 3.5)

## Documentation

Golang

- https://go.dev/

Golang-Migrate

- https://github.com/golang-migrate/migrate
- To handle migrations with Golang
- Install:
  - curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
  - sudo mv migrate.linux-amd64 $GOPATH/bin/migrate

SQLC

- https://sqlc.dev/
- To compile SQL to type-safe code
- How to use:
  - Create the queries in a .sql file and insert `-- name: MethodName :exec | :many | :one`
  - Run: `sqlc generate`
  - It will use the sqlc.yaml to create new files in "internal/infra/db/\*"

OpenAI

- https://github.com/sashabaranov/go-openai
- OpenAI API

## gRPC

It's a framework and a communication protocal created by Google, which can help us to work with system integration.
It's using HTTP2 and Protocol Buffer, and can work with Scream.
Documentation: https://grpc.io/docs/what-is-grpc/

- Comunication prococal contract: proto/chat.proto
- Installing proto-c: https://grpc.io/docs/protoc-installation/

## Getting Started

1. Run docker compose: `docker-compose up -d`
2. Run the migrations: `make migrate`
3. Secret Key:

- Generate the Secret Key (here)[https://platform.openai.com/account/api-keys]

4. Test:

```
docker-compose exec mysql bash

bash-4.4# mysql -root -p chat_test;
Enter password: root

mysql> show tables;

+---------------------+
| Tables_in_chat_test |
+---------------------+
| chats               |
| messages            |
| schema_migrations   |
+---------------------+
```

5. Build TikToken inside the container. Para gerar o arquivo libtiktoken.a

```
docker-compose exec chatservice bash

root@59d7ae6490be:/go/src# cd tiktoken-cffi/
cargo build --release
cp target/release/libtiktoken.a .
```

6. Run: `root@59d7ae6490be:/go/src# go run cmd/chatservice/main.go`
