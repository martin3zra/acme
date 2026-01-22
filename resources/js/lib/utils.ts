import { PaymentMethod, PaymentMethods } from '@/types';
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

export const getNetDays = (term: string): number => {
  const match = term.match(/^Net\s*(\d+)$/i);
  if (match) {
    return parseInt(match[1], 10);
  }
  return 0; // Return null if not a standard "Net" term
};

export const isPaymentMethod = (value: string): value is PaymentMethod => {
  return (PaymentMethods as readonly string[]).includes(value);
};

export const parsePaymentMethod = (value: string, fallback: PaymentMethod = 'cash'): PaymentMethod => {
  return isPaymentMethod(value) ? value : fallback;
};

export const capitalize = (s: string) => (s.length ? s[0].toUpperCase() + s.substring(1) : s);
