import { useTranslation } from '@/hooks/use-translation';
import { LegendProps } from 'recharts';

export function CustomLegend({ payload }: LegendProps) {
  const t = useTranslation().trans;
  if (!payload) return null; // guard against undefined
  const filtered = payload.filter((p) => p.dataKey !== 'salesFill');
  return (
    <ul className="flex flex-col space-y-1">
      {filtered.map((entry, index) => (
        <li key={`item-${index}`} className="flex items-center text-xs font-medium text-gray-500">
          <span className="mr-2 h-3 w-3 rounded" style={{ backgroundColor: entry.color }} />
          <span className="capitalize">{t(`dashboard.chart.legend.${entry.value}`)}</span>
        </li>
      ))}
    </ul>
  );
}
