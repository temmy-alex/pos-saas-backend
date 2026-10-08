# Store Status API

API ini mengelola register kas harian per branch dan kompatibel dengan endpoint POS mobile SuperPOS.

Semua endpoint membutuhkan Bearer token.

## Get status

```http
GET /api/store/status?branch_id=1&date=2026-10-08
```

`date` opsional dan menggunakan tanggal hari ini di zona waktu `Asia/Jakarta`. Response berisi status register, total penjualan POS, pembagian cash/transfer, dan expected cash.

## Open store

```http
POST /api/store/open
Content-Type: application/json
```

```json
{
  "branch_id": 1,
  "cash_open": 500000,
  "closed_at": "22:00"
}
```

`closed_at` bersifat opsional dan menunjukkan rencana waktu tutup. Register yang sudah terbuka atau sudah ditutup pada tanggal yang sama tidak dapat dibuka ulang.

## Close store

```http
POST /api/store/close
Content-Type: application/json
```

```json
{
  "branch_id": 1,
  "cash_close": 1250000
}
```

Saat closing, API menyimpan:

- `expected_cash_at_close`: saldo awal ditambah penjualan cash bersih.
- `cash_difference`: saldo fisik saat closing dikurangi expected cash.
- `closed_at`: jam closing dalam zona waktu Jakarta.
- `closed_at_timestamp`: timestamp lengkap untuk audit.

Cashier dapat mengosongkan `branch_id`; API akan memakai branch dari JWT. Admin dan superadmin harus mengirim `branch_id` jika token tidak mempunyai default branch.

## Transaction guard

`POST /api/transactions` hanya dapat membuat transaksi ketika register branch untuk tanggal berjalan sedang terbuka. Row register dikunci selama pembuatan transaksi sehingga checkout yang sedang berjalan dan proses closing tidak dapat menghasilkan perhitungan kas yang saling mendahului.
