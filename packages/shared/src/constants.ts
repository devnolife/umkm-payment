/**
 * Shared label/color/business-logic constants.
 * Renderer-agnostic: hex codes (mobile) and Tailwind classes (web).
 */
import type { OrderStatus, PaymentMethod, PaymentStatus, UserRole } from './types';

export const ORDER_STATUS_LABEL: Record<OrderStatus, string> = {
  PENDING: 'Menunggu Konfirmasi',
  CONFIRMED: 'Dikonfirmasi',
  PROCESSING: 'Sedang Diproses',
  READY: 'Siap Diambil',
  COMPLETED: 'Selesai',
  CANCELLED: 'Dibatalkan',
};

export const ORDER_STATUS_COLOR: Record<OrderStatus, string> = {
  PENDING: '#F59E0B',
  CONFIRMED: '#3B82F6',
  PROCESSING: '#8B5CF6',
  READY: '#10B981',
  COMPLETED: '#6B7280',
  CANCELLED: '#EF4444',
};

export const ORDER_STATUS_BADGE_CLASS: Record<OrderStatus, string> = {
  PENDING: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  CONFIRMED: 'bg-blue-100 text-blue-800 border-blue-200',
  PROCESSING: 'bg-violet-100 text-violet-800 border-violet-200',
  READY: 'bg-emerald-100 text-emerald-800 border-emerald-200',
  COMPLETED: 'bg-gray-100 text-gray-600 border-gray-200',
  CANCELLED: 'bg-red-100 text-red-700 border-red-200',
};

export const PAYMENT_METHOD_LABEL: Record<PaymentMethod, string> = {
  COD: 'Bayar di Tempat (COD)',
  ONLINE: 'Pembayaran Online',
};

export const PAYMENT_STATUS_LABEL: Record<PaymentStatus, string> = {
  UNPAID: 'Belum Bayar',
  PENDING: 'Menunggu Pembayaran',
  PAID: 'Sudah Dibayar',
  FAILED: 'Gagal',
  REFUNDED: 'Dikembalikan',
};

export const PAYMENT_STATUS_BADGE_CLASS: Record<PaymentStatus, string> = {
  UNPAID: 'bg-gray-100 text-gray-600 border-gray-200',
  PENDING: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  PAID: 'bg-green-100 text-green-700 border-green-200',
  FAILED: 'bg-red-100 text-red-700 border-red-200',
  REFUNDED: 'bg-blue-100 text-blue-700 border-blue-200',
};

export const USER_ROLE_LABEL: Record<UserRole, string> = {
  BUYER: 'Pembeli',
  SELLER: 'Penjual',
  ADMIN: 'Admin',
};

export const FOOD_CATEGORIES = [
  { key: 'all', label: 'Semua', icon: '🍽️' },
  { key: 'rice', label: 'Nasi', icon: '🍚' },
  { key: 'noodle', label: 'Mie', icon: '🍜' },
  { key: 'snack', label: 'Jajanan', icon: '🍢' },
  { key: 'drink', label: 'Minuman', icon: '🥤' },
  { key: 'dessert', label: 'Dessert', icon: '🍰' },
  { key: 'coffee', label: 'Kopi', icon: '☕' },
] as const;
export type FoodCategoryKey = (typeof FOOD_CATEGORIES)[number]['key'];

// ==================== ORDER STATUS FLOW ====================

const ORDER_FLOW: OrderStatus[] = [
  'PENDING',
  'CONFIRMED',
  'PROCESSING',
  'READY',
  'COMPLETED',
];

export function getNextOrderStatus(current: OrderStatus): OrderStatus | null {
  if (current === 'CANCELLED' || current === 'COMPLETED') return null;
  const idx = ORDER_FLOW.indexOf(current);
  if (idx === -1 || idx >= ORDER_FLOW.length - 1) return null;
  return ORDER_FLOW[idx + 1];
}

export function getNextOrderStatusLabel(current: OrderStatus): string | null {
  const next = getNextOrderStatus(current);
  if (!next) return null;
  const actionLabels: Partial<Record<OrderStatus, string>> = {
    CONFIRMED: 'Konfirmasi',
    PROCESSING: 'Proses',
    READY: 'Siap Diambil',
    COMPLETED: 'Selesai',
  };
  return actionLabels[next] ?? null;
}

export function canCancelOrder(status: OrderStatus): boolean {
  return status === 'PENDING';
}

export function isActiveOrder(status: OrderStatus): boolean {
  return !(['COMPLETED', 'CANCELLED'] as OrderStatus[]).includes(status);
}

export function getOrderProgressPercent(status: OrderStatus): number {
  const map: Record<OrderStatus, number> = {
    PENDING: 10,
    CONFIRMED: 30,
    PROCESSING: 55,
    READY: 80,
    COMPLETED: 100,
    CANCELLED: 0,
  };
  return map[status] ?? 0;
}
