/**
 * Shared domain types & enums for UMKM platform.
 * Values mirror the Prisma schema in apps/web/prisma/schema.prisma.
 */

// ==================== ENUMS ====================

export const UserRole = {
  BUYER: 'BUYER',
  SELLER: 'SELLER',
  ADMIN: 'ADMIN',
} as const;
export type UserRole = (typeof UserRole)[keyof typeof UserRole];

export const OrderStatus = {
  PENDING: 'PENDING',
  CONFIRMED: 'CONFIRMED',
  PROCESSING: 'PROCESSING',
  READY: 'READY',
  COMPLETED: 'COMPLETED',
  CANCELLED: 'CANCELLED',
} as const;
export type OrderStatus = (typeof OrderStatus)[keyof typeof OrderStatus];

export const PaymentMethod = {
  COD: 'COD',
  ONLINE: 'ONLINE',
} as const;
export type PaymentMethod = (typeof PaymentMethod)[keyof typeof PaymentMethod];

export const PaymentStatus = {
  UNPAID: 'UNPAID',
  PENDING: 'PENDING',
  PAID: 'PAID',
  FAILED: 'FAILED',
  REFUNDED: 'REFUNDED',
} as const;
export type PaymentStatus = (typeof PaymentStatus)[keyof typeof PaymentStatus];

// ==================== DOMAIN MODELS ====================

export interface User {
  id: string;
  username: string;
  name: string;
  email: string | null;
  phone: string | null;
  role: UserRole;
  avatar: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Store {
  id: string;
  sellerId: string;
  name: string;
  description: string | null;
  address: string;
  latitude: number | null;
  longitude: number | null;
  phone: string;
  image: string | null;
  isOpen: boolean;
  isVerified: boolean;
  openTime: string;
  closeTime: string;
  createdAt: string;
  updatedAt: string;
}

export interface Category {
  id: string;
  storeId: string;
  name: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface MenuItem {
  id: string;
  storeId: string;
  categoryId: string | null;
  name: string;
  description: string | null;
  price: number;
  image: string | null;
  isAvailable: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface OrderItem {
  id: string;
  orderId: string;
  menuItemId: string;
  quantity: number;
  price: number;
  notes: string | null;
  menuItem?: MenuItem;
}

export interface Order {
  id: string;
  orderNumber: string;
  buyerId: string;
  storeId: string;
  status: OrderStatus;
  totalPrice: number;
  paymentMethod: PaymentMethod;
  paymentStatus: PaymentStatus;
  notes: string | null;
  estimatedReadyTime: string | null;
  createdAt: string;
  updatedAt: string;
  orderItems?: OrderItem[];
  buyer?: User;
  store?: Store;
  payment?: Payment;
}

export interface Payment {
  id: string;
  orderId: string;
  method: string;
  amount: number;
  status: PaymentStatus;
  midtransTransactionId: string | null;
  midtransSnapToken: string | null;
  midtransRedirectUrl: string | null;
  paidAt: string | null;
  createdAt: string;
  updatedAt: string;
}

// ==================== API CONTRACTS ====================

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface AuthSession {
  user: Pick<User, 'id' | 'username' | 'name' | 'role' | 'avatar'>;
  token: string;
  expiresAt: string;
}
