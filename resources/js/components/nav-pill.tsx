import { cn } from '@/lib/utils';
// ── Pill Label ────────────────────────────────────────────────────────────────
const pillStyles: Record<string, string> = {
  soon: 'bg-muted text-muted-foreground',
  new: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  beta: 'bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-400',
  pro: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
};

export default function NavPill({ label, variant = 'soon' }: { label: string; variant?: string }) {
  return <span className={cn('ml-auto rounded-full px-2 py-0.5 text-[11px] leading-none font-medium', pillStyles[variant])}>{label}</span>;
}
