# Customer API

Semua endpoint membutuhkan header:

```http
Authorization: Bearer {access_token}
Accept: application/json
```

## Mobile-compatible endpoints

Endpoint berikut mempertahankan kontrak yang digunakan POS mobile SuperPOS.

### List customers

```http
GET /api/master/customers?page=1&search=alex&branch_id=1
```

Query opsional:

- `page`: halaman, default `1`.
- `search`: mencari nama, telepon, atau email.
- `branch_id`: filter cabang yang dapat diakses user.
- `store_id`: filter store untuk superadmin.
- `per_page` atau `limit`: jumlah data, default `10`, maksimal `100`.

Response:

```json
{
  "message": "Data customer berhasil diambil",
  "data": [],
  "pagination": {
    "page": 1,
    "per_page": 10,
    "total": 0,
    "has_more": false
  }
}
```

### Create customer

```http
POST /api/master/customers
Content-Type: application/json
```

```json
{
  "name": "Alex",
  "phone": "081234567890",
  "email": "alex@example.com",
  "branch_id": 1
}
```

`branch_id` boleh tidak dikirim oleh cashier karena otomatis memakai branch dari token. Admin dan superadmin wajib mengirimkannya jika token tidak memiliki default branch.

## Canonical endpoints

API backend juga menyediakan endpoint lengkap dengan format response standar project:

```text
GET    /api/customers
POST   /api/customers
GET    /api/customers/:id
PUT    /api/customers/:id
DELETE /api/customers/:id
```

Cashier dapat melihat dan membuat customer di branch miliknya. Update dan delete hanya tersedia untuk admin dan superadmin. Semua operasi menerapkan scope store dan branch dari JWT.
