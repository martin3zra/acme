import { useTranslation } from '@/hooks/use-translation';
import { TooltipProps } from 'recharts';

interface PayloadEntry {
  dataKey: string;
  name: string;
  value: number;
  color: string;
}

interface CustomTooltipProps extends TooltipProps<number, string> {
  payload?: PayloadEntry[];
  label?: string;
  formatter?: (value: number) => string;
}

export const tooltipContentStyle = {
  backgroundColor: '#f9fafb',
  borderRadius: '0.5rem',
  border: '1px solid #e5e7eb',
  fontSize: '12px',
  fontWeight: 500,
  color: '#374151',
  padding: '8px 12px',
};

export const tooltipLabelStyle = {
  fontSize: '12px',
  fontWeight: 600,
  color: '#111827',
};

export const CustomTooltip: React.FC<CustomTooltipProps> = ({ payload, label, formatter }) => {
  const t = useTranslation().trans;
  if (label === undefined || label === null) return null;
  if (!payload || payload.length === 0) return null;
  // Split into [year, monthNumber]
  const [year, monthStr] = label.split('/');
  if (!year || !monthStr) return '';
  const filtered = payload.filter((p) => p.dataKey !== 'salesFill');
  return (
    <div className="rounded border bg-gray-50 p-2 text-xs">
      <div className="mb-1 font-semibold text-gray-800">
        {year}/{t(`dashboard.chart.months.${monthStr.toLowerCase()}`)}
      </div>
      {filtered.map((entry, i) => (
        <div key={i} style={{ color: entry.color }}>
          <span className="capitalize">{t(`dashboard.chart.legend.${entry.name}`)}</span>: {formatter ? formatter(entry.value) : entry.value}
        </div>
      ))}
    </div>
  );
};
