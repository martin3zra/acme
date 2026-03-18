export default function NavBadge({ count }: { count: number }) {
  if (count <= 0) return null;
  return (
    <span className="bg-primary text-primary-foreground ring-background ml-auto flex h-5 min-w-5 items-center justify-center rounded-full px-1.5 text-[11px] font-semibold tabular-nums ring-1">
      {count > 99 ? '99+' : count}
    </span>
  );
}
