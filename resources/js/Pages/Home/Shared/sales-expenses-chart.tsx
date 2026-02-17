import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import { ChartPoint, Replacements, Totals } from '@/types';
import { Area, ComposedChart, Legend, Line, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { EmptyChartState } from './empty-state';
import { CustomLegend } from './legend-content';
import { CustomTooltip, tooltipContentStyle, tooltipLabelStyle } from './tooltip-content';

interface Props {
  period: string;
  chartData: ChartPoint[];
  totals: Totals;
  availableYears: number[]; // e.g. [2024, 2025]
  onPeriodChange: (newPeriod: string) => void;
}

export default function SalesExpensesChart({ period, chartData, totals, availableYears, onPeriodChange }: Props) {
  const { currency, abbrvNumber } = useNumber();
  const t = useTranslation().trans;
  // duplicate the sales data for the area fill
  chartData = chartData.map((d) => ({ ...d, salesFill: d.sales }));
  const isEmptyTotals = Object.values({
    sales: totals.totalSales,
    receipts: totals.totalReceipts,
    expenses: totals.totalExpenses,
  }).every((val) => !val);

  return (
    <div className="rounded-lg p-6 shadow">
      <div className="flex w-full justify-between">
        <div className="fex basis-4/5 flex-col">
          <div className="flex w-full items-center justify-between pb-4">
            <h2 className="text-lg font-semibold text-gray-800">{t('dashboard.chart.title')}</h2>
            <select
              value={period}
              onChange={(e) => onPeriodChange(e.target.value)}
              className="rounded-md border border-gray-300 px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
            >
              <option value="last12">{t('dashboard.chart.last12Months')}</option>
              {availableYears.map((year) => (
                <option key={year} value={`year-${year}`}>
                  {year}
                </option>
              ))}
            </select>
          </div>
          {isEmptyTotals ? (
            <EmptyChartState />
          ) : (
            <div className="h-64 flex-1">
              <ResponsiveContainer width="100%" height="100%">
                <ComposedChart data={chartData}>
                  <defs>
                    <linearGradient id="salesAreaGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.4} />
                      <stop offset="100%" stopColor="#3b82f6" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <XAxis dataKey="month" tick={{ fontSize: 12, fill: '#6b7280', fontWeight: 500 }} tickFormatter={(label) => formatLabel(label, t)} />
                  <YAxis
                    tick={{ fontSize: 12, fill: '#6b7280', fontWeight: 500 }}
                    tickFormatter={(value) => (value === 0 ? '' : abbrvNumber(value))}
                  />
                  <Tooltip contentStyle={tooltipContentStyle} labelStyle={tooltipLabelStyle} content={<CustomTooltip formatter={currency} />} />
                  <Area type="monotone" dataKey="salesFill" stroke="none" fill="url(#salesAreaGradient)" fillOpacity={1} />
                  <Line type="monotone" dataKey="sales" stroke="#3b82f6" strokeWidth={3} dot={{ r: 4 }} />
                  <Line type="monotone" dataKey="expenses" stroke="#ef4444" strokeWidth={3} dot={{ r: 4 }} />
                  <Legend verticalAlign="middle" align="right" layout="vertical" content={<CustomLegend />} />
                </ComposedChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>
        {isEmptyTotals ? null : (
          <div className="flex basis-auto flex-col lg:flex-row">
            <div className="flex w-full flex-col items-end space-y-4 border-l lg:w-64 [&_div]:flex [&_div]:flex-col [&_div]:items-end">
              <div>
                <p className="text-sm text-gray-500">{t('dashboard.chart.summary.sales')}</p>
                <p className="text-lg font-bold text-gray-800">{currency(totals.totalSales)}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">{t('dashboard.chart.summary.receipts')}</p>
                <p className="text-lg font-bold text-gray-800">{currency(totals.totalReceipts)}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">{t('dashboard.chart.summary.expenses')}</p>
                <p className="text-lg font-bold text-gray-800">{currency(totals.totalExpenses)}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">{t('dashboard.chart.summary.netIncome')}</p>{' '}
                <p className="text-lg font-bold text-gray-800">{currency(totals.netIncome)}</p>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function formatLabel(label: string | undefined, trans: (key: string, replacements?: Replacements) => string) {
  if (!label) return '';

  const [year, monthStr] = label.split('/');
  if (!year || !monthStr) return '';

  const tranlatedMonth = trans(`dashboard.chart.months.${monthStr.toLowerCase()}`);

  return `${year}/${tranlatedMonth}`;
}
