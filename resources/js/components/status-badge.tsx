import type { StatusType } from '@/types/index';
import { statusConfig } from '@/types/status-config-map';
import React from 'react';

interface StatusBadgeProps {
  type: StatusType;
  status: string;
  className?: string;
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ type, status, className = '' }) => {
  const config = statusConfig[type][status];

  if (!config) return null;

  const { Icon, label, bg, text } = config;

  return (
    <div className={`inline-flex items-center gap-2 rounded-full px-3 py-1 text-sm font-medium ${bg} ${text} ${className}`}>
      <Icon className={`h-4 w-4 ${text}`} />
      <span>{label}</span>
    </div>
  );
};
