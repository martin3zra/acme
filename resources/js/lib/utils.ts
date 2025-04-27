import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export  const addDays = (dateValue: Date, days: number): Date => {
  const date = new Date(dateValue.valueOf())
  date.setDate(date.getDate() + days)
  return date
}