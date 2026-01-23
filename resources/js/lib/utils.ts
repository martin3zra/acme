import { PaymentMethod, PaymentMethods, PaymentTermValue } from '@/types';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const addDays = (dateValue: Date, days: number): Date => {
  const date = new Date(dateValue.valueOf());
  date.setDate(date.getDate() + days);
  return date;
};

export const isNotEmpty = (value: string | null | undefined): boolean => {
  return !!value && value.trim().length > 0;
};

export const subtractFloats = (a: number, b: number, precision: number = 2): number => {
  const result = Math.abs(a - b);
  const factor = Math.pow(10, precision);
  return Math.round(result * factor) / factor;
};

export const isPaymentMethod = (value: string): value is PaymentMethod => {
  return (PaymentMethods as readonly string[]).includes(value);
};

export const parsePaymentMethod = (value: string, fallback: PaymentMethod = 'cash'): PaymentMethod => {
  return isPaymentMethod(value) ? value : fallback;
};

export const capitalize = (s: string) => (s.length ? s[0].toUpperCase() + s.substring(1) : s);

export function getDaysFromTerm(term: PaymentTermValue): number {
  if (term === 'pia') return 0 as any;
  const match = term.match(/^net(\d+)$/);
  return match ? parseInt(match[1], 10) : 0;
}
