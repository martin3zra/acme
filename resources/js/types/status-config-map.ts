import type { StatusType } from '@/types/index';
import {
  AlertTriangle,
  CheckCircle,
  CheckCircle2,
  CircleDashed,
  CircleDollarSign,
  Clock,
  Eye,
  FileCheck,
  FileText,
  LucideIcon,
  Send,
  ShieldBan,
  ShieldCheck,
  Slash,
  Trash2,
  Wallet,
  XCircle,
} from 'lucide-react';

type StatusConfig = {
  label: string;
  bg: string;
  border: string;
  text: string;
  Icon: LucideIcon;
};

export const statusConfig: Record<StatusType, Record<string, StatusConfig>> = {
  paid: {
    paid: {
      label: 'Paid',
      bg: 'bg-green-100',
      border: 'border-green-500',
      text: 'text-green-800',
      Icon: CheckCircle,
    },
    unpaid: {
      label: 'Unpaid',
      bg: 'bg-red-100',
      border: 'border-red-500',
      text: 'text-red-800',
      Icon: XCircle,
    },
    partial: {
      label: 'Partial',
      bg: 'bg-yellow-100',
      border: 'border-yellow-500',
      text: 'text-yellow-800',
      Icon: CircleDashed,
    },
    removed: {
      label: 'Removed',
      bg: 'bg-gray-200',
      border: 'border-gray-500',
      text: 'text-gray-700',
      Icon: Trash2,
    },
    overpaid: {
      label: 'Overpaid',
      bg: 'bg-blue-100',
      border: 'border-blue-500',
      text: 'text-blue-800',
      Icon: CircleDollarSign,
    },
    pending: {
      label: 'Pending',
      bg: 'bg-orange-100',
      border: 'border-orange-500',
      text: 'text-orange-800',
      Icon: Clock,
    },
  },
  invoice: {
    open: {
      label: 'Open',
      bg: 'bg-yellow-100',
      border: 'border-yellow-500',
      text: 'text-yellow-700',
      Icon: FileCheck,
    },
    draft: {
      label: 'Draft',
      bg: 'bg-gray-100',
      border: 'border-gray-500',
      text: 'text-gray-700',
      Icon: FileText,
    },
    sent: {
      label: 'Sent',
      bg: 'bg-blue-100',
      border: 'border-blue-500',
      text: 'text-blue-800',
      Icon: Send,
    },
    viewed: {
      label: 'Viewed',
      bg: 'bg-indigo-100',
      border: 'border-indigo-500',
      text: 'text-indigo-800',
      Icon: Eye,
    },
    overdue: {
      label: 'Overdue',
      bg: 'bg-orange-100',
      border: 'border-orange-500',
      text: 'text-orange-800',
      Icon: AlertTriangle,
    },
    completed: {
      label: 'Completed',
      bg: 'bg-green-100',
      border: 'border-green-500',
      text: 'text-green-800',
      Icon: CheckCircle2,
    },
    void: {
      label: 'Void',
      bg: 'bg-red-100',
      border: 'border-red-500',
      text: 'text-red-800',
      Icon: Slash,
    },
    partial: {
      label: 'Partial',
      bg: 'bg-teal-100',
      border: 'border-teal-500',
      text: 'text-teal-800',
      Icon: Wallet,
    },
  },
  status: {
    enabled: {
      label: 'Enabled',
      bg: 'bg-green-100',
      border: 'border-green-500',
      text: 'text-green-800',
      Icon: ShieldCheck,
    },
    disabled: {
      label: 'Disabled',
      bg: 'bg-gray-200',
      border: 'border-gray-500',
      text: 'text-gray-700',
      Icon: ShieldBan,
    },
  },
};
