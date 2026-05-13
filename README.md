# UMKM Payment

Platform pemesanan makanan untuk UMKM. Monorepo berisi web admin/seller/buyer (Next.js) dan mobile app (Expo / React Native).

## Tech Stack

| Layer       | Tech                                                                                  |
| ----------- | ------------------------------------------------------------------------------------- |
| Monorepo    | pnpm workspaces · Turborepo                                                           |
| Backend API | **Go 1.22 · Fiber v2 · GORM** (REST + WebSocket)                                      |
| Web         | Next.js 16 (App Router) · React 19 · TypeScript · Tailwind v4 · shadcn/ui (frontend)  |
| Mobile      | Expo 54 · React Native 0.81 · React Navigation 7 · Zustand                            |
| Auth        | JWT (HS256, bcrypt) — shared antara mobile & web                                      |
| Database    | PostgreSQL · GORM AutoMigrate (skema didefinisikan di `apps/api/internal/models`) |
| Realtime    | gofiber/websocket (native WebSocket)                                                  |
| Payments    | Midtrans (Snap, QRIS, bank transfer, e-wallet) — handled by Go API                    |
| Validation  | go-playground/validator (Go) · Zod (TS) di `@umkm/shared`                             |

## Repository Layout

```
umkm/
├── apps/
│   ├── api/          # Go (Fiber + GORM) — REST + WebSocket backend
│   ├── web/          # Next.js — admin/buyer/seller frontend (konsumsi @umkm/api)
│   └── mobile/       # Expo — buyer & seller app (konsumsi @umkm/api)
├── packages/
│   └── shared/       # @umkm/shared — types, enums, Zod schemas, formatters
├── turbo.json
└── pnpm-workspace.yaml
```

> **Catatan arsitektur:** Semua endpoint REST & WebSocket dilayani oleh `apps/api` (Go). Web & mobile mengonsumsi `apps/api` via `NEXT_PUBLIC_API_URL` / config mobile. Skema database adalah model GORM di `apps/api/internal/models/models.go`; tabel dibuat otomatis lewat `AutoMigrate` saat API server start dengan `APP_ENV=development`.

## Prerequisites

- Node.js ≥ 20
- pnpm 10 (`npm i -g pnpm@10`)
- **Go ≥ 1.22** (`https://go.dev/dl/`)
- PostgreSQL ≥ 14 (lokal atau remote)
- Expo CLI tidak perlu install global (pakai `pnpm dev:mobile`)

## Setup

```powershell
# 1. Install dependencies
pnpm install

# 2. Konfigurasi environment
Copy-Item apps\web\.env apps\web\.env.local  # edit nilai sesuai kebutuhan
```

Isi minimum `apps/web/.env`:

```env
DATABASE_URL="postgresql://user:pass@localhost:5432/umkm_db?schema=public"
NEXTAUTH_URL="http://localhost:3000"
NEXTAUTH_SECRET="<generate dengan: openssl rand -base64 32>"
JWT_SECRET="<generate dengan: openssl rand -base64 32>"
JWT_EXPIRES_IN="7d"
MIDTRANS_SERVER_KEY="<dari dashboard Midtrans>"
MIDTRANS_CLIENT_KEY="<dari dashboard Midtrans>"
MIDTRANS_IS_PRODUCTION=false
NEXT_PUBLIC_APP_URL="http://localhost:3000"
NEXT_PUBLIC_WS_URL="http://localhost:3001"
```

```powershell
# 3. Migrate database (otomatis saat API start dengan APP_ENV=development)
#    & seed akun + toko + menu contoh
pnpm dev:api         # biarkan jalan sebentar agar AutoMigrate jalan, lalu Ctrl+C
pnpm api:seed        # bisa dijalankan berulang (idempotent)
pnpm api:dbcheck     # opsional: cek koneksi DB & row counts
```

### Akun seed default

Setelah `pnpm api:seed`, akun-akun berikut siap dipakai (di **web** maupun **mobile** — auth shared via JWT):

| Role     | Username     | Password    | Keterangan                                  |
| -------- | ------------ | ----------- | ------------------------------------------- |
| ADMIN    | `admin`      | `admin123`  | Dashboard admin web (`/admin`)              |
| SELLER   | `warungbudi` | `seller123` | Pemilik toko **Warung Budi** (`/seller`)    |
| SELLER   | `kantinmaya` | `seller123` | Pemilik toko **Kantin Maya** (`/seller`)    |
| BUYER    | `pembeli1`   | `buyer123`  | Buyer untuk uji flow web (`/buyer`)         |
| BUYER    | `pembeli2`   | `buyer123`  | Buyer kedua                                 |
| BUYER    | `mobileuser` | `mobile123` | Buyer khusus untuk app Expo (mobile)        |

Seed bersifat **idempotent**: re-running tidak menggandakan data dan tidak menimpa password user yang sudah ada (hanya nama/role/email yang diperbarui).

## Scripts (root)

| Script            | Aksi                                                      |
| ----------------- | --------------------------------------------------------- |
| `pnpm dev`        | Jalankan semua app JS paralel (web + mobile)              |
| `pnpm dev:web`    | Hanya web Next.js (port 3000)                             |
| `pnpm dev:mobile` | Hanya Expo dev server                                     |
| `pnpm dev:api`    | **Jalankan Go API server (port 4000)** — `go run`         |
| `pnpm api:tidy`   | `go mod tidy` di `apps/api`                               |
| `pnpm api:build`  | Build binary Go ke `apps/api/bin/api.exe`                 |
| `pnpm api:seed`   | Seed akun admin/seller/buyer + toko & menu contoh         |
| `pnpm api:dbcheck`| Cek koneksi DB & jumlah row tiap tabel                    |
| `pnpm build`      | Build semua JS app via Turborepo                          |
| `pnpm lint`       | Lint semua JS app                                         |

## Domain Model (ringkas)

```
User (BUYER | SELLER | ADMIN)
  └─ Store (1:1 dengan SELLER)
       ├─ Category (n)
       └─ MenuItem (n)
            └─ OrderItem ──▶ Order ──▶ Payment
```

- **Harga** disimpan sebagai `Int` (Rupiah, tanpa desimal).
- **OrderItem.price** adalah snapshot harga saat order dibuat.
- **Order status flow:** PENDING → CONFIRMED → PROCESSING → READY → COMPLETED (atau CANCELLED).

## Design System (UI Pattern)

**Arah visual:** *Friendly & Appetizing* — warm orange (food-app vibe). Mirror dari mobile palette di `apps/mobile/src/theme/colors.js`.

### Design tokens (web)
Didefinisikan via CSS variables di `apps/web/src/app/globals.css` dan diekspos ke Tailwind v4. Pakai kelas, jangan hardcode warna.

| Token                 | Kegunaan                                            |
| --------------------- | --------------------------------------------------- |
| `bg-primary`          | Tombol utama, CTA (= warna brand)                   |
| `bg-brand` / `text-brand` | Aksen brand eksplisit (sama dengan primary)     |
| `bg-brand-soft`       | Background lembut untuk badge / icon wrap           |
| `bg-success / -soft`  | Status sukses (emerald)                             |
| `bg-warning / -soft`  | Status warning (amber)                              |
| `bg-info / -soft`     | Status informasi (blue)                             |
| `bg-destructive`      | Error / aksi destruktif (red)                       |
| `font-display`        | Heading / wordmark (Plus Jakarta Sans)              |
| `font-sans` (default) | Body teks (Inter)                                   |
| `bg-brand-gradient`   | Background gradient marketing / hero                |
| `shadow-brand`        | Drop shadow ber-tint brand                          |

### Pattern components (`@/components/patterns`)
Reusable building blocks — pakai daripada inline berulang.

| Komponen      | Kegunaan                                                              |
| ------------- | --------------------------------------------------------------------- |
| `BrandLogo`   | Logo Jajanin (3 varian: solid / soft / onBrand, 3 ukuran).            |
| `PageHeader`  | Heading konsisten setiap page dashboard (title + description + actions). |
| `StatCard`    | KPI card dengan icon, value, tone (brand/success/info/warning/neutral), optional trend. |
| `EmptyState`  | Zero-state untuk list kosong (icon + title + description + CTA).      |
| `Section`     | Content section dengan optional heading + trailing action.            |

Contoh:
```tsx
import { PageHeader, StatCard, EmptyState } from '@/components/patterns';
import { ShoppingBag } from 'lucide-react';

<PageHeader
  eyebrow="Dashboard"
  title="Halo, Budi 👋"
  description="Ringkasan toko Anda hari ini."
/>
<StatCard tone="brand" icon={ShoppingBag} label="Pesanan" value={12} />
<EmptyState icon={ShoppingBag} title="Belum ada pesanan" />
```

### Konsistensi web ↔ mobile
Mobile sudah punya design token system lengkap di `apps/mobile/src/theme/`. Web token di `globals.css` adalah port langsung dari sana (orange brand, slate text, emerald/amber/red/blue/violet status). Jika ada perubahan brand, **ubah keduanya**.

## Shared Code (`@umkm/shared`)

Single source of truth untuk types, enum, Zod schema, dan formatter yang dipakai web & mobile. Lihat [packages/shared/README.md](packages/shared/README.md).

```ts
import { CreateOrderInputSchema, ORDER_STATUS_LABEL, formatRupiah } from '@umkm/shared';
```

## Catatan Keamanan

- File `.env` di `apps/web/` dan `apps/api/` **tidak di-commit** (sudah ada di `.gitignore`). Hanya `.env.example` / `.env.local.example` yang boleh masuk repo, isinya placeholder saja.
- Selalu pakai `.env.local` (web) atau `.env` lokal (api) untuk secret nyata; jangan push.
- Rotate `NEXTAUTH_SECRET` dan `JWT_SECRET` sebelum deploy production. Kalau secret pernah ada di working tree yang ke-share, anggap bocor dan rotate.
- Production CORS: set `CORS_ORIGINS` di `apps/api/.env` ke allowlist eksplisit (mis. `https://app.example.com`). Nilai `*` ditolak saat `APP_ENV != development` dan akan menggagalkan startup.
- WebSocket (`/ws`) butuh JWT lewat query string: `wss://host/ws?token=<JWT>`. Token dengan signature/expiry invalid akan otomatis ditolak handshake.

## Token Refresh

Access token (JWT) default berumur 7 hari (`JWT_EXPIRES_HOURS=168`). Untuk rotasi tanpa logout, panggil `POST /api/auth/refresh` dengan `Authorization: Bearer <current-token>` selagi token masih valid. Response: `{ "data": { "token": "<new-token>" } }`. Token yang sudah expired tidak bisa di-refresh — user harus login ulang. Untuk environment yang lebih ketat, turunkan `JWT_EXPIRES_HOURS` (mis. 1 jam) dan biasakan client memanggil `/auth/refresh` secara berkala.

## Lisensi

Proprietary — internal use only.
