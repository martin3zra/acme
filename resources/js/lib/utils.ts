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
  const result = a - b;
  const factor = Math.pow(10, precision);
  return Math.round(result * factor) / factor;
};
