import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import type { StatusType, TransactionKind } from '@/types/index';
import { statusConfig } from '@/types/status-config-map';
import React from 'react';

interface StatusBadgeProps {
  kind?: TransactionKind;
  type: StatusType;
  status: string;
  className?: string;
  variant?: 'badge' | 'alert';
  prefix?: string;
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ kind, type, status, className = '', variant = 'badge', prefix }) => {
  const t = useTranslation().trans;
  const config = statusConfig[type][status];

  if (!config) return null;

  const { Icon, bg, border, text } = config;

  const resolveLabel = (): string => {
    let path = 'invoice';
    if (kind) {
      path = kind;
    }
    return {
      invoice: t(path + 's.statuses.' + status),
      status: t('global.statuses.' + status),
      paid: t('global.paidStatuses.' + status),
      payment: t('global.paidStatuses.' + status),
      dashboard: t('dashboard.statuses.' + status),
    }[type];
  };

  return (
    <div
      className={cn(
        `inline-flex items-center gap-2 text-sm font-medium ${bg} ${text} ${className}`,
        variant === 'badge' && 'rounded-full px-3 py-1',
        variant === 'alert' && `rounded-md border px-4 py-3 ${border}`,
      )}
    >
      <Icon className={`h-4 w-4 ${text}`} />
      <span>
        {prefix && <span className="pe-1">{prefix}</span>}
        {resolveLabel()}
      </span>
    </div>
  );
};
