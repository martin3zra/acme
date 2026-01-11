import AppLayout from '@/layouts/app-layout';
import { DueInvoice, PageProps, StatItem } from '@/types';
import { router } from '@inertiajs/react';
import { DueInvoicesList } from './DueInvoicesList';
import DashboardSummary from './Shared/dashboard-summary';

interface DashboardProps extends PageProps {
  stats: StatItem[];
  due_invoices: DueInvoice[];
}

export default function Home({ auth, stats, due_invoices }: DashboardProps) {
  const handleOnSelectItem = (item: DueInvoice, action: 'view:customer' | 'view:invoice' | 'record:payment') => {
    const path = {
      'view:customer': `/customers?id=${item.customer.uuid}`,
      'view:invoice': `/invoices?id=${item.uuid}`,
      'record:payment': `/payments/create?customer_id=${item.customer.uuid}&invoice_id=${item.uuid}`,
    }[action];
    router.visit(path);
  };
  return (
    <AppLayout user={auth.user}>
      <div className="space-y-6">
        <DashboardSummary stats={stats} />
        <div className="grid auto-rows-min gap-4 sm:grid-cols-1 md:grid-cols-2">
          <DueInvoicesList data={due_invoices} onSelectItem={handleOnSelectItem} />
          <div className="bg-muted/50 rounded-xl" />
        </div>
      </div>
    </AppLayout>
  );
}
