import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { BreadcrumbItem, Invoice, PageProps } from '@/types';
import { Link } from '@inertiajs/react';
import { Plus } from 'lucide-react';

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

  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasInvoices && (
          <HeadingSmall
            title="Invoices"
            description="All created invoices are shown here"
            rightPanel={
              <Button>
                <Plus /> Add Invoice
              </Button>
            }
          />
        )}

        {!hasInvoices && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">Create your first invoice</h4>
              <p className="text-sm text-gray-400">Once you create your invoice, it will appear here.</p>
              <Link
                href="/invoices/create"
                as="button"
                className="focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive bg-primary text-primary-foreground hover:bg-primary/90 inline-flex h-9 shrink-0 cursor-pointer items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium whitespace-nowrap shadow-xs transition-all outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 has-[>svg]:px-3 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4"
              >
                + Create Invoice
              </Link>
            </div>
          </>
        )}
      </div>
    </AuthenticatedLayout>
  );
}
