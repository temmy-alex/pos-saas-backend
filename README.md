# POS SaaS Backend

Backend API untuk project POS SaaS Multi Branch menggunakan Golang, PostgreSQL, dan Docker.

## Tech Stack

- Golang
- Gin Framework
- PostgreSQL
- Docker Compose

## Struktur Folder

```text
pos-saas-backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── routes/
│   ├── services/
│   ├── repositories/
│   ├── models/
│   ├── middlewares/
│   └── helpers/
├── migrations/
├── docker-compose.yml
├── .env.example
├── go.mod
└── README.md