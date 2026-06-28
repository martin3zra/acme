import { cn } from '@/lib/utils';
import type { PillVariant } from '@/types/index';

// ── Count Badge ──────────────────────────────────────────────
interface NavBadgeProps {
  count: number;
  className?: string;
}

export function NavBadge({ count, className }: NavBadgeProps) {
  if (count <= 0) return null;

  return (
    <span
      className={cn(
        'ml-auto flex h-5 min-w-5 items-center justify-center rounded-full',
        'bg-primary text-primary-foreground px-1.5 text-[11px] font-semibold tabular-nums',
        'ring-background ring-1',
        className,
      )}
    >
      {count > 99 ? '99+' : count}
    </span>
  );
}

// ── Pill Label ────────────────────────────────────────────────
const pillVariantStyles: Record<PillVariant, string> = {
  soon: 'bg-muted text-muted-foreground',
  new: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  beta: 'bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-400',
  pro: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
};

interface NavPillProps {
  label: string;
  variant?: PillVariant;
  className?: string;
}

export function NavPill({ label, variant = 'soon', className }: NavPillProps) {
  return (
    <span className={cn('ml-auto rounded-full px-2 py-0.5 text-[11px] leading-none font-medium', pillVariantStyles[variant], className)}>
      {label}
    </span>
  );
}
