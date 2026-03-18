import type { StatusType } from '@/types/index';
import {
  AlertTriangle,
  Archive,
  CheckCircle,
  CheckCircle2,
  CircleCheckBig,
  CircleDashed,
  CircleSlash,
  Clock,
  FileText,
  LucideIcon,
  Package,
  RotateCcw,
  Send,
  ShieldBan,
  ShieldCheck,
  Slash,
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
    unpaid: {
      label: 'Unpaid',
      bg: 'bg-red-100',
      border: 'border-red-500',
      text: 'text-red-800',
      Icon: XCircle,
    },
    partial: {
      label: 'Partial',
      bg: 'bg-teal-100',
      border: 'border-teal-500',
      text: 'text-teal-800',
      Icon: CircleDashed,
    },
    paid: {
      label: 'Paid',
      bg: 'bg-green-100',
      border: 'border-green-500',
      text: 'text-green-800',
      Icon: CheckCircle,
    },
    refund: {
      label: 'Refund',
      bg: 'bg-purple-100',
      border: 'border-purple-500',
      text: 'text-purple-800',
      Icon: RotateCcw,
    },
  },
  invoice: {
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
    overdue: {
      label: 'Overdue',
      bg: 'bg-orange-100',
      border: 'border-orange-500',
      text: 'text-orange-800',
      Icon: AlertTriangle,
    },
    void: {
      label: 'Void',
      bg: 'bg-red-100',
      border: 'border-red-500',
      text: 'text-red-800',
      Icon: Slash,
    },
    closed: {
      label: 'Closed',
      bg: 'bg-green-100',
      border: 'border-green-500',
      text: 'text-green-800',
      Icon: CheckCircle2,
    },
    uncollectible: {
      label: 'Uncollectible',
      bg: 'bg-purple-100',
      border: 'border-purple-500',
      text: 'text-purple-800',
      Icon: Archive,
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
  payment: {
    pending: {
      label: 'Pending',
      bg: 'bg-yellow-100',
      border: 'border-yellow-500',
      text: 'text-yellow-800',
      Icon: Clock,
    },
    completed: {
      label: 'Completed',
      bg: 'bg-blue-100',
      border: 'border-blue-500',
      text: 'text-blue-800',
      Icon: CircleCheckBig,
    },
    failed: {
      label: 'Failed',
      bg: 'bg-red-100',
      border: 'border-red-500',
      text: 'text-red-800',
      Icon: XCircle,
    },
    void: {
      label: 'Voided',
      bg: 'bg-gray-100',
      border: 'border-gray-400',
      text: 'text-gray-700',
      Icon: CircleSlash,
    },
  },
  dashboard: {
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
    closed: {
      label: 'Closed',
      bg: 'bg-green-100',
      border: 'border-green-500',
      text: 'text-green-800',
      Icon: CheckCircle2,
    },
  },
  purchase: {
    draft: {
      label: 'Draft',
      bg: 'bg-gray-100',
      border: 'border-gray-500',
      text: 'text-gray-700',
      Icon: FileText,
    },
    open: {
      label: 'Open',
      bg: 'bg-gray-100',
      border: 'border-gray-500',
      text: 'text-gray-700',
      Icon: Package,
    },
    received: {
      label: 'Received',
      bg: 'bg-green-100',
      border: 'border-green-500',
      text: 'text-green-800',
      Icon: CheckCircle2,
    },
    partially_received: {
      label: 'Partially Received',
      bg: 'bg-teal-100',
      border: 'border-teal-500',
      text: 'text-teal-800',
      Icon: CircleDashed,
    },
  },
};
