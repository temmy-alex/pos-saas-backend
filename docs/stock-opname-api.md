# Stock Opname API

Semua endpoint membutuhkan Bearer token dan menerapkan scope store/branch dari JWT.

## Mobile-compatible list

```http
GET /api/master/stock-opnames?branch_id=1&page=1&search=10
```

Response memakai bentuk detail yang kompatibel dengan SuperPOS: setiap item mempunyai data `product`, `stock_opname.branch`, dan `stock_opname.user`.

## List and detail

```http
GET /api/stock-opnames?branch_id=1&status=draft&page=1&limit=10
GET /api/stock-opnames/:id
```

Nilai `status` yang tersedia adalah `draft`, `completed`, dan `cancelled`.

## Create draft

```http
POST /api/stock-opnames
Content-Type: application/json
```

```json
{
  "branch_id": 1,
  "note": "Opname akhir bulan",
  "items": [
    {
      "product_id": 10,
      "old_stock": 12,
      "new_stock": 10,
      "note": "Dua barang rusak"
    }
  ]
}
```

`old_stock` wajib dikirim sebagai optimistic concurrency check. Jika stok database sudah berubah, API mengembalikan `409 Conflict` dan draft tidak dibuat.

Payload legacy `stockOpnameInput` juga diterima sebagai pengganti `items`.

## Complete

```http
POST /api/stock-opnames/:id/complete
```

Completion mengunci opname dan seluruh produk, memeriksa ulang `old_stock`, lalu memperbarui stok secara atomik. Jika satu produk conflict, tidak ada stok yang berubah.

## Cancel

```http
POST /api/stock-opnames/:id/cancel
Content-Type: application/json
```

```json
{
  "reason": "Perhitungan fisik harus diulang"
}
```

Hanya draft yang dapat diselesaikan atau dibatalkan. Opname completed tidak dapat dibatalkan karena stoknya sudah menjadi bagian dari audit inventory.
