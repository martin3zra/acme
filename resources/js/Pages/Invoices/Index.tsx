import HeadingSmall from '@/components/heading-small';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { BreadcrumbItem, Invoice, PageProps, Verb } from '@/types';
import { List } from './List/Index';
import { AddNewInvoice } from './Shared/AddNewInvoice';

const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Home',
    href: '/home',
  },
  {
    title: 'Invoices',
    href: '/invoices',
  },
];
export default function Index({ auth, invoices }: PageProps<{ invoices: Invoice[] }>) {
  const hasInvoices = invoices.length > 0;

  const onSelectInvoice = (invoice: Invoice, action: Verb): void => {}

  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasInvoices && (
          <HeadingSmall
            title="Invoices"
            description="All created invoices are shown here"
            rightPanel={<AddNewInvoice />}
          />
        )}

        {!hasInvoices && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">Create your first invoice</h4>
              <p className="text-sm text-gray-400">Once you create your invoice, it will appear here.</p>
              <AddNewInvoice />
            </div>
          </>
        )}

        {hasInvoices && <List data={invoices} onSelectInvoice={onSelectInvoice} />}
      </div>
    </AuthenticatedLayout>
  );
}
