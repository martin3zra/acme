import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { ChartPoint, DueInvoice, PageProps, StatItem, Totals } from '@/types';
import { router } from '@inertiajs/react';
import { useState } from 'react';
import { RecentTransactionList } from './RecentTransactionList';
import DashboardSummary from './Shared/dashboard-summary';
import SalesExpensesChart from './Shared/sales-expenses-chart';
import { ProgressProps, WelcomeBoard } from './Shared/welcome-board';

interface ChartDataPoint {
  data: ChartPoint[];
  totals: Totals;
  availableYears: number[];
}

interface DashboardProps extends PageProps {
  hasMissingData: boolean;
  progress: ProgressProps;
  stats: StatItem[];
  period: string;
  due_invoices: DueInvoice[];
  estimates: DueInvoice[];
  chart: ChartDataPoint;
}

export default function Home({ auth, hasMissingData, progress, stats, period: initialPeriod, due_invoices, estimates, chart }: DashboardProps) {
  const t = useTranslation().trans;
  const [period, setPeriod] = useState<string>(initialPeriod);
  const handleOnSelectItem = (item: DueInvoice, action: 'view:customer' | 'view:invoice' | 'view:estimate' | 'record:payment') => {
    const path = {
      'view:customer': `/customers?id=${item.customer.uuid}`,
      'view:invoice': `/invoices?id=${item.uuid}`,
      'view:estimate': `/estimates?id=${item.uuid}`,
      'record:payment': `/payments/create?customer_id=${item.customer.uuid}&invoice_id=${item.uuid}`,
    }[action];
    router.visit(path);
  };
  const handleOnPeriodChange = (newPeriod: string) => {
    setPeriod(newPeriod);
    router.visit('/home', {
      except: ['stats', 'due_invoices'],
      data: { period: newPeriod },
      preserveState: true,
      preserveScroll: true,
    });
  };
  return (
    <AppLayout user={auth.user}>
      {hasMissingData ? (
        <div className="mb-4">
          <WelcomeBoard progress={progress} />
        </div>
      ) : (
        <div className="space-y-6">
          <DashboardSummary stats={stats} />
          <div className="py-4">
            <SalesExpensesChart
              period={period}
              chartData={chart.data}
              totals={chart.totals}
              availableYears={chart.availableYears}
              onPeriodChange={handleOnPeriodChange}
            />
          </div>
          <div className="grid auto-rows-min gap-4 sm:grid-cols-1 md:grid-cols-2">
            <RecentTransactionList title={t('dashboard.due_invoices')} kind={'invoice'} data={due_invoices} onSelectItem={handleOnSelectItem} />
            <RecentTransactionList title={t('dashboard.estimates')} kind={'estimate'} data={estimates} onSelectItem={handleOnSelectItem} />
          </div>
        </div>
      )}
    </AppLayout> 
  );
}
