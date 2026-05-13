# @umkm/api — Go Backend Service

Backend REST + WebSocket untuk UMKM Payment, dibangun dengan **Go + Fiber + GORM + PostgreSQL**. Menggantikan API routes di `apps/web` (Next.js) — web Next.js sekarang fokus jadi frontend (admin, buyer, seller) yang konsumsi service Go ini.

## Tech Stack

| Layer       | Tech                                              |
| ----------- | ------------------------------------------------- |
| Language    | Go 1.22+                                          |
| HTTP        | [Fiber v2](https://gofiber.io)                    |
| WebSocket   | gofiber/websocket                                 |
| ORM         | GORM (PostgreSQL driver)                          |
| Auth        | JWT (golang-jwt/jwt v5) + bcrypt                  |
| Validation  | go-playground/validator                           |
| Payment     | Midtrans Snap (HTTP client manual)                |

## Struktur

```
apps/api/
├── cmd/server/main.go              # entrypoint
├── internal/
│   ├── config/                     # env loader
│   ├── database/                   # GORM connection + auto-migrate
│   ├── models/                     # GORM models (mirror Prisma schema)
│   ├── dto/                        # request DTO + validator tags
│   ├── middleware/                 # JWT auth + role guard
│   ├── services/                   # jwt, midtrans
│   ├── handlers/                   # HTTP handlers + routes.go
│   ├── ws/                         # websocket hub + handler
│   └── utils/                      # response, password, validator, id, order#
├── .env.example
└── go.mod
```

## Setup

```powershell
# 1. Install Go 1.22+ → https://go.dev/dl/

# 2. Masuk ke folder
cd apps\api

# 3. Copy env & isi nilainya
Copy-Item .env.example .env

# 4. Pastikan DATABASE_URL menunjuk ke database PostgreSQL yang sama dengan Prisma
#    Schema dikelola Prisma (apps/web). Jalankan dulu:
#    pnpm db:push  (dari root)

# 5. Install dependencies
go mod tidy

# 6. Run
go run ./cmd/server
```

Server akan listen di `http://localhost:4000`.

> **Catatan migrasi:** Prisma di `apps/web` tetap sumber kebenaran skema. Service Go ini hanya **konsumen** database. `AutoMigrate` di dev hanya untuk kenyamanan; di production matikan via `APP_ENV=production`.

## Environment Variables

Lihat `.env.example`. Variabel wajib:

- `DATABASE_URL` — PostgreSQL connection string
- `JWT_SECRET` — secret HMAC (generate: `openssl rand -base64 32`)
- `MIDTRANS_SERVER_KEY` — dari dashboard Midtrans (sandbox/production)

## API Endpoints

Semua endpoint di-prefix `/api`.

### Auth
| Method | Path                | Auth          |
| ------ | ------------------- | ------------- |
| POST   | `/auth/register`    | public        |
| POST   | `/auth/login`       | public        |
| GET    | `/auth/profile`     | JWT           |
| PUT    | `/auth/profile`     | JWT           |

### Stores
| Method | Path                          | Auth          |
| ------ | ----------------------------- | ------------- |
| GET    | `/stores`                     | public        |
| GET    | `/stores/:id`                 | public        |
| GET    | `/stores/:id/menu`            | public        |
| GET    | `/stores/:id/queue?orderId=`  | public        |
| POST   | `/stores`                     | SELLER        |
| PUT    | `/stores/:id`                 | SELLER/ADMIN  |
| PATCH  | `/stores/:id/toggle`          | SELLER/ADMIN  |

### Menu & Categories
| Method | Path                | Auth          |
| ------ | ------------------- | ------------- |
| POST   | `/categories`       | SELLER/ADMIN  |
| PUT    | `/categories/:id`   | SELLER/ADMIN  |
| DELETE | `/categories/:id`   | SELLER/ADMIN  |
| POST   | `/menu`             | SELLER/ADMIN  |
| PUT    | `/menu/:id`         | SELLER/ADMIN  |
| DELETE | `/menu/:id`         | SELLER/ADMIN  |

### Orders
| Method | Path                         | Auth                 |
| ------ | ---------------------------- | -------------------- |
| POST   | `/orders`                    | BUYER                |
| GET    | `/orders`                    | JWT (auto-scoped)    |
| GET    | `/orders/:id`                | JWT                  |
| PATCH  | `/orders/:id/status`         | JWT (role-checked)   |

Status flow: `PENDING → CONFIRMED → PROCESSING → READY → COMPLETED` (atau `CANCELLED`).
Buyer hanya boleh cancel order PENDING miliknya. Seller boleh transisi status untuk store miliknya.

### Payments
| Method | Path                  | Auth     | Catatan                              |
| ------ | --------------------- | -------- | ------------------------------------ |
| POST   | `/payments/create`    | BUYER    | Buat Snap token Midtrans             |
| POST   | `/payments/webhook`   | public   | Notif Midtrans (verifikasi SHA512)   |

### Admin
| Method | Path                            | Auth   |
| ------ | ------------------------------- | ------ |
| GET    | `/admin/stats`                  | ADMIN  |
| GET    | `/admin/users`                  | ADMIN  |
| PATCH  | `/admin/users/:id/toggle`       | ADMIN  |

## WebSocket

Endpoint: `ws://localhost:4000/ws?token=<JWT>`

Protokol pesan client → server (JSON):
```json
{"type":"join","room":"order:abc123"}
{"type":"leave","room":"order:abc123"}
{"type":"ping"}
```

Room yang otomatis di-join:
- `user:<userId>` — saat connect

Room yang bisa di-join manual:
- `store:<storeId>` — seller mendengarkan order baru di store-nya
- `order:<orderId>` — buyer / seller tracking realtime per-order

Event yang dipancarkan server:
- `order.created` — order baru dibuat
- `order.status` — status order berubah
- `payment.update` — status pembayaran berubah

## Integrasi dengan Mobile

Update `apps/mobile/src/constants/index.js`:

```js
export const API_URL = 'http://10.0.2.2:4000';  // Android emulator
// atau IP LAN PC untuk device fisik: 'http://192.168.x.x:4000'
```

Lalu di `apps/mobile/src/services/index.js`, set:

```js
const USE_MOCK_API = false;
```

Untuk Socket: gunakan native WebSocket (bukan socket.io-client) karena backend pakai gofiber/websocket:
```js
const ws = new WebSocket(`ws://10.0.2.2:4000/ws?token=${token}`);
```

## Build untuk Production

```powershell
go build -o bin/api.exe ./cmd/server
.\bin\api.exe
```

Set `APP_ENV=production` agar AutoMigrate tidak ikut jalan.
