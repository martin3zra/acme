import { Months, NavItem, PaymentMethod, PaymentMethods, PaymentTermValue, Recurrent } from '@/types';
import { clsx, type ClassValue } from 'clsx';
import { formatDate } from 'date-fns';
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
  if (term === 'pia') return 0;
  const match = term.match(/^net(\d+)$/);
  return match ? parseInt(match[1], 10) : 0;
}

export function calculateDueDate(terms: PaymentTermValue, issueDate: Date = new Date()): Date | undefined {
  switch (terms) {
    case 'pia': // Payable Immediately After
      return issueDate; // same day
    case 'net7':
      return addDays(issueDate, 7);
    case 'net15':
      return addDays(issueDate, 15);
    case 'net30':
      return addDays(issueDate, 30);
    case 'net60':
      return addDays(issueDate, 60);
    case 'net90':
      return addDays(issueDate, 90);
    default:
      return undefined; // no due date
  }
}

export function validateRecurrence(data: Recurrent): Recurrent {
  const base: Recurrent = {
    enabled: data.enabled,
    name: data.name,
    type: data.type,
    frequency: data.frequency,
    interval: Math.max(1, data.interval), // always >= 1
    start_date: data.start_date,
    until: data.until,
    timezone: data.timezone,
    send_email: data.send_email,
  };

  if (data.start_date && data.until) {
    const start = new Date(data.start_date);
    const until = new Date(data.until);
    if (start > until) {
      throw new Error('Start date cannot be greater than until date.');
    }
  }
  switch (data.frequency) {
    case 'daily':
      if (base.interval < 1) {
        throw new Error('Daily recurrence requires an interval of at least 1 day.');
      }
      return base;
    case 'weekly':
      if (!data.weekdays || data.weekdays.length === 0) {
        throw new Error('Weekly recurrence requires at least one weekday selected.');
      }
      if (base.interval > 52) {
        throw new Error('Weekly recurrence interval cannot exceed 52 weeks.');
      }
      return { ...base, weekdays: data.weekdays };
    case 'monthly':
      if (!data.day_of_month) {
        throw new Error('Monthly recurrence requires a valid day of the month.');
      }
      if (base.interval > 12) {
        throw new Error('Monthly recurrence interval cannot exceed 12 months.');
      }
      return { ...base, day_of_month: clampDay(data.day_of_month) };
    case 'quarterly':
      if (!data.day_of_month) {
        throw new Error('Quarterly recurrence requires a valid day of the month.');
      }
      if (base.interval > 4) {
        throw new Error('Quarterly recurrence interval cannot exceed 4 quarters.');
      }
      return { ...base, day_of_month: clampDay(data.day_of_month) };
    case 'yearly':
      if (!data.month) {
        throw new Error('Yearly recurrence requires a month to be selected.');
      }
      if (!data.day_of_month) {
        throw new Error('Yearly recurrence requires a valid day of the month.');
      }
      if (base.interval > 10) {
        throw new Error('Yearly recurrence interval cannot exceed 10 years.');
      }
      return { ...base, month: clampMonth(data.month), day_of_month: clampDay(data.day_of_month) };
  }
}

// Helpers
function clampDay(day: number): number {
  if (day < 1) return 1;
  if (day > 31) return 31;
  return day;
}

function clampMonth(month: number): number {
  if (month < 1) return 1;
  if (month > 12) return 12;
  return month;
}

export function formatOrdinal(n: number): string {
  const s = ['th', 'st', 'nd', 'rd'];
  const v = n % 100;
  return n + (s[(v - 20) % 10] || s[v] || s[0]);
}

export function formatWeekdays(weekdays: string[]): string {
  if (!weekdays || weekdays.length === 0) return '';
  if (weekdays.length === 1) return weekdays[0];
  return weekdays.slice(0, -1).join(', ') + ' and ' + weekdays[weekdays.length - 1];
}

export function recurrenceCaption(data: Recurrent): string {
  let caption = 'Repeat ';

  switch (data.frequency) {
    case 'daily':
      caption += `every ${data.interval} day${data.interval > 1 ? 's' : ''}`;
      break;

    case 'weekly':
      caption += `every ${data.interval} week${data.interval > 1 ? 's' : ''}`;
      if (data.weekdays && data.weekdays.length > 0) {
        caption += ` on ${formatWeekdays(data.weekdays)}`;
      }
      break;

    case 'monthly':
      caption += `on ${formatOrdinal(data.day_of_month || 0)} day of every ${data.interval} month${data.interval > 1 ? 's' : ''}`;
      break;

    case 'quarterly':
      caption += `on ${formatOrdinal(data.day_of_month || 0)} day of every ${data.interval} quarter${data.interval > 1 ? 's' : ''}`;
      break;

    case 'yearly':
      if (data.month && data.day_of_month) {
        caption += `on ${formatOrdinal(data.day_of_month || 0)} day of ${Months[data.month - 1]} every ${data.interval} year${data.interval > 1 ? 's' : ''}`;
      } else {
        caption += `every ${data.interval} year${data.interval > 1 ? 's' : ''}`;
      }
      break;
  }

  if (data.start_date) {
    caption += ` starting ${formatDate(data.start_date, 'dd-MM-yyyy')}`;
  }

  if (data.until) {
    caption += ` until ${formatDate(data.until, 'dd-MM-yyyy')}`;
  }

  return caption;
}

export function parseLocalDate(dateStr: string) {
  const [y, m, d] = dateStr.split('-').map(Number);
  return new Date(y, m - 1, d);
}

export function clampISODate(value: string | undefined | null): string {
  if (!value) return '';
  return value.split('T')[0]; // "2026-02-27"
}

const groupMap: Record<string, string> = {
  'global.navMain.dashboard': 'global.navGroups.overview',

  // Sales
  'global.navMain.invoices': 'global.navGroups.sales',
  'global.navMain.estimates': 'global.navGroups.sales',
  'global.navMain.orders': 'global.navGroups.sales',

  // Purchasing
  'global.navMain.purchaseOrders': 'global.navGroups.purchasing',
  'global.navMain.purchaseReceipts': 'global.navGroups.purchasing',
  'global.navMain.vendorBills': 'global.navGroups.purchasing',

  // Inventory
  'global.navMain.warehouses': 'global.navGroups.inventory',
  'global.navMain.stock': 'global.navGroups.inventory',
  'global.navMain.transfers': 'global.navGroups.inventory',
  'global.navMain.adjustments': 'global.navGroups.inventory',

  // Catalog
  'global.navMain.customers': 'global.navGroups.customersCatalog',
  'global.navMain.vendors': 'global.navGroups.customersCatalog',
  'global.navMain.items': 'global.navGroups.customersCatalog',
  'global.navMain.attributes': 'global.navGroups.customersCatalog',

  // Finance
  'global.navMain.receivables': 'global.navGroups.finance',
  'global.navMain.payables': 'global.navGroups.finance',
  'global.navMain.expenses': 'global.navGroups.finance',

  // Analytics
  'global.navMain.reports': 'global.navGroups.analytics',
};

export function buildNavGroups(navMain: NavItem[]) {
  const groups: { group: string; items: NavItem[] }[] = [];

  navMain.forEach((item) => {
    const groupKey = groupMap[item.title];
    if (!groupKey) return;

    let group = groups.find((g) => g.group === groupKey);
    if (!group) {
      group = { group: groupKey, items: [] };
      groups.push(group);
    }
    group.items.push(item);
  });

  return groups;
}
