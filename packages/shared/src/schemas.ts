/**
 * Shared Zod validation schemas. Single source of truth for input validation
 * on the web API and mobile client. Use `z.infer<typeof X>` to derive types.
 */
import { z } from 'zod';

// ==================== ENUMS ====================

export const UserRoleSchema = z.enum(['BUYER', 'SELLER', 'ADMIN']);
export const OrderStatusSchema = z.enum([
  'PENDING',
  'CONFIRMED',
  'PROCESSING',
  'READY',
  'COMPLETED',
  'CANCELLED',
]);
export const PaymentMethodSchema = z.enum(['COD', 'ONLINE']);
export const PaymentStatusSchema = z.enum([
  'UNPAID',
  'PENDING',
  'PAID',
  'FAILED',
  'REFUNDED',
]);

// ==================== AUTH ====================

export const LoginInputSchema = z.object({
  username: z.string().min(3, 'Username minimal 3 karakter'),
  password: z.string().min(6, 'Password minimal 6 karakter'),
});
export type LoginInput = z.infer<typeof LoginInputSchema>;

export const RegisterInputSchema = z.object({
  username: z
    .string()
    .min(3, 'Username minimal 3 karakter')
    .max(30)
    .regex(/^[a-zA-Z0-9_]+$/, 'Username hanya huruf, angka, dan garis bawah'),
  name: z.string().min(2, 'Nama minimal 2 karakter').max(100),
  email: z.string().email('Email tidak valid').optional().or(z.literal('')),
  phone: z
    .string()
    .regex(/^(\+62|62|0)8[1-9][0-9]{6,11}$/, 'Nomor HP tidak valid')
    .optional()
    .or(z.literal('')),
  password: z.string().min(6, 'Password minimal 6 karakter').max(72),
  role: UserRoleSchema.optional().default('BUYER'),
});
export type RegisterInput = z.infer<typeof RegisterInputSchema>;

// ==================== STORE ====================

const TimeStringSchema = z
  .string()
  .regex(/^([01]\d|2[0-3]):[0-5]\d$/, 'Format waktu harus HH:mm');

export const StoreInputSchema = z.object({
  name: z.string().min(2).max(100),
  description: z.string().max(500).optional().nullable(),
  address: z.string().min(5).max(255),
  latitude: z.number().min(-90).max(90).optional().nullable(),
  longitude: z.number().min(-180).max(180).optional().nullable(),
  phone: z.string().regex(/^(\+62|62|0)8[1-9][0-9]{6,11}$/, 'Nomor HP tidak valid'),
  image: z.string().url().optional().nullable(),
  openTime: TimeStringSchema.default('08:00'),
  closeTime: TimeStringSchema.default('21:00'),
});
export type StoreInput = z.infer<typeof StoreInputSchema>;

// ==================== CATEGORY ====================

export const CategoryInputSchema = z.object({
  name: z.string().min(1).max(50),
  sortOrder: z.number().int().min(0).default(0),
});
export type CategoryInput = z.infer<typeof CategoryInputSchema>;

// ==================== MENU ITEM ====================

export const MenuItemInputSchema = z.object({
  categoryId: z.string().cuid().optional().nullable(),
  name: z.string().min(2).max(100),
  description: z.string().max(500).optional().nullable(),
  price: z.number().int().min(0, 'Harga tidak boleh negatif'),
  image: z.string().url().optional().nullable(),
  isAvailable: z.boolean().default(true),
});
export type MenuItemInput = z.infer<typeof MenuItemInputSchema>;

// ==================== ORDER ====================

export const OrderItemInputSchema = z.object({
  menuItemId: z.string().cuid(),
  quantity: z.number().int().min(1).max(99),
  notes: z.string().max(255).optional().nullable(),
});
export type OrderItemInput = z.infer<typeof OrderItemInputSchema>;

export const CreateOrderInputSchema = z.object({
  storeId: z.string().cuid(),
  items: z.array(OrderItemInputSchema).min(1, 'Minimal 1 item'),
  paymentMethod: PaymentMethodSchema.default('COD'),
  notes: z.string().max(500).optional().nullable(),
});
export type CreateOrderInput = z.infer<typeof CreateOrderInputSchema>;

export const UpdateOrderStatusSchema = z.object({
  status: OrderStatusSchema,
  estimatedReadyTime: z.string().datetime().optional().nullable(),
});
export type UpdateOrderStatusInput = z.infer<typeof UpdateOrderStatusSchema>;

// ==================== PAYMENT ====================

export const PaymentInitInputSchema = z.object({
  orderId: z.string().cuid(),
  method: z.string().min(1),
});
export type PaymentInitInput = z.infer<typeof PaymentInitInputSchema>;
