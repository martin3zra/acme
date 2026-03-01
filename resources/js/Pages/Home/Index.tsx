import { EmptyState } from '@/components/empty-state';
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
            {!progress.invoices_created ? (
              <EmptyState
                title="Aún no has creado ninguna factura"
                description="Crea tu primera factura para empezar a registrar tus ventas."
                variant="action"
                actionLabel="+ Crear mi primera factura"
                onAction={() => router.visit('/invoices/create')}
              />
            ) : progress.invoices_created && chart.data.length === 0 ? (
              <EmptyState
                title="Todavía no hay datos gráficos disponibles"
                description="Cierra o registra facturas para ver tus ventas y gastos aquí."
                variant="action"
                actionLabel="+ Crear factura"
                onAction={() => router.visit('/invoices/create')}
              />
            ) : (
              <SalesExpensesChart
                period={period}
                chartData={chart.data}
                totals={chart.totals}
                availableYears={chart.availableYears}
                onPeriodChange={handleOnPeriodChange}
              />
            )}
          </div>
          <div className="grid auto-rows-min gap-4 sm:grid-cols-1 md:grid-cols-2">
            {due_invoices.length === 0 ? (
              progress.invoices_created ? (
                <EmptyState title="Sin facturas vencidas 🎉" description="Tus clientes están al día con sus pagos." variant="positive" />
              ) : (
                <EmptyState
                  title="Sin facturas"
                  description="Aún no has creado ninguna factura."
                  variant="action"
                  actionLabel="Crear factura"
                  onAction={() => router.visit('/invoices/create')}
                />
              )
            ) : (
              <RecentTransactionList title={t('dashboard.due_invoices')} kind={'invoice'} data={due_invoices} onSelectItem={handleOnSelectItem} />
            )}

            {estimates.length === 0 ? (
              progress.estimates_created ? (
                <EmptyState
                  title="No hay cotizaciones recientes"
                  description="Revisa tu historial para ver cotizaciones anteriores."
                  variant="positive"
                />
              ) : (
                <EmptyState
                  title="Aún no has creado ninguna cotización"
                  description="Genera tu primera cotización para empezar a cerrar negocios."
                  variant="action"
                  actionLabel="+ Nueva cotización"
                  onAction={() => router.visit('/estimates/create')}
                />
              )
            ) : (
              <RecentTransactionList title={t('dashboard.estimates')} kind={'estimate'} data={estimates} onSelectItem={handleOnSelectItem} />
            )}
          </div>
        </div>
      )}
    </AppLayout>
  );
}
