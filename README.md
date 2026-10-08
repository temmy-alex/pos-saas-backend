# POS SaaS Backend - Golang + PostgreSQL + Docker + Goose + JWT

POS SaaS Backend adalah project backend POS berbasis Golang yang dibangun untuk kebutuhan pembelajaran real project.

Project ini menggunakan:

- Golang
- Gin Web Framework
- PostgreSQL
- Docker Compose
- Goose Migration
- JWT Authentication
- Role Based Access
- Local Image Upload
- Transaction API
- Daily Sales Report
- Dashboard API
- Receipt API

Project ini cocok untuk belajar membangun backend aplikasi POS yang lebih realistis, bukan hanya CRUD sederhana.

---

## 1. Fitur Utama

### Auth

- Login menggunakan email dan password
- Generate JWT access token
- Protected route menggunakan Bearer token
- Endpoint authenticated user

### Role Access

Role yang tersedia:

- superadmin
- admin
- cashier

Aturan akses data:

- superadmin dapat mengakses semua store dan branch
- admin hanya dapat mengakses data store miliknya
- cashier hanya dapat mengakses data branch miliknya

### Master Data

- Store CRUD
- Branch CRUD
- Category CRUD
- Product CRUD
- Upload image product secara local
- Product filter, search, dan pagination
- Customer CRUD, search, pagination, dan mobile-compatible API
- Store open/close, daily cash register, dan cash reconciliation
- Stock opname draft, approval, stock adjustment, dan audit trail

### POS Transaction

- Create transaction
- List transaction
- Detail transaction
- Stock product otomatis berkurang saat transaksi
- Void transaction
- Stock product otomatis kembali saat transaksi di-void
- Receipt API untuk struk

### Report & Dashboard

- Daily sales report
- Payment breakdown
- Top products
- Dashboard summary
- Sales per hour
- Low stock products
- Recent transactions

---

## 2. Struktur Project

```text
pos-saas-backend/
├── cmd/
│   └── api/
│       └── main.go
├── docs/
│   └── API_TESTING.md
├── internal/
│   ├── auth/
│   │   └── claims.go
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── helpers/
│   ├── middlewares/
│   ├── models/
│   ├── repositories/
│   ├── requests/
│   ├── responses/
│   ├── routes/
│   └── services/
├── migrations/
├── uploads/
│   └── products/
├── docker-compose.yml
├── go.mod
├── go.sum
├── .env.example
└── README.md# POS SaaS Backend - Golang + PostgreSQL + Docker + Goose + JWT

POS SaaS Backend adalah project backend POS berbasis Golang yang dibangun untuk kebutuhan pembelajaran real project.

Project ini menggunakan:

- Golang
- Gin Web Framework
- PostgreSQL
- Docker Compose
- Goose Migration
- JWT Authentication
- Role Based Access
- Local Image Upload
- Transaction API
- Daily Sales Report
- Dashboard API
- Receipt API

Project ini cocok untuk belajar membangun backend aplikasi POS yang lebih realistis, bukan hanya CRUD sederhana.

---

## 1. Fitur Utama

### Auth

- Login menggunakan email dan password
- Generate JWT access token
- Protected route menggunakan Bearer token
- Endpoint authenticated user

### Role Access

Role yang tersedia:

- superadmin
- admin
- cashier

Aturan akses data:

- superadmin dapat mengakses semua store dan branch
- admin hanya dapat mengakses data store miliknya
- cashier hanya dapat mengakses data branch miliknya

### Master Data

- Store CRUD
- Branch CRUD
- Category CRUD
- Product CRUD
- Upload image product secara local
- Product filter, search, dan pagination

### POS Transaction

- Create transaction
- List transaction
- Detail transaction
- Stock product otomatis berkurang saat transaksi
- Void transaction
- Stock product otomatis kembali saat transaksi di-void
- Receipt API untuk struk

### Report & Dashboard

- Daily sales report
- Payment breakdown
- Top products
- Dashboard summary
- Sales per hour
- Low stock products
- Recent transactions

---

## 2. Struktur Project

```text
pos-saas-backend/
├── cmd/
│   └── api/
│       └── main.go
├── docs/
│   └── API_TESTING.md
├── internal/
│   ├── auth/
│   │   └── claims.go
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── helpers/
│   ├── middlewares/
│   ├── models/
│   ├── repositories/
│   ├── requests/
│   ├── responses/
│   ├── routes/
│   └── services/
├── migrations/
├── uploads/
│   └── products/
├── docker-compose.yml
├── go.mod
├── go.sum
├── .env.example
└── README.md
