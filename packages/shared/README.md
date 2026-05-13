# @umkm/shared

Shared types, enums, constants, Zod validation schemas, and formatting helpers used across the UMKM monorepo.

## Modules

| Path                  | Contents                                                          |
| --------------------- | ----------------------------------------------------------------- |
| `@umkm/shared/types`     | Domain models & enums mirroring `prisma/schema.prisma`            |
| `@umkm/shared/constants` | Labels, colors, badge classes, order-flow helpers                 |
| `@umkm/shared/schemas`   | Zod validation schemas (auth, store, menu, order, payment)        |
| `@umkm/shared/format`    | Renderer-agnostic formatters (Rupiah, dates, greetings, initials) |

The barrel export `@umkm/shared` re-exports everything.

## Usage

### In `@umkm/web`
```ts
import { CreateOrderInputSchema, formatRupiah, ORDER_STATUS_LABEL } from '@umkm/shared';

const parsed = CreateOrderInputSchema.parse(await req.json());
```

### In `@umkm/mobile`
```js
import { formatRupiah, ORDER_STATUS_LABEL } from '@umkm/shared';
```

## Adding the dependency

Add to the consumer app's `package.json`:
```json
"dependencies": { "@umkm/shared": "workspace:*" }
```
Then run `pnpm install` at the repo root.

## Conventions

- This package owns the **single source of truth** for cross-app contracts. When the Prisma schema changes, update `src/types.ts` and `src/schemas.ts` in lockstep.
- Keep this package free of platform-specific imports (no `next/*`, no `react-native`).
- Prices are integers (Rupiah, no decimals) — never use floats.
