import HeadingSmall from '@/components/heading-small';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { BreadcrumbItem, Invoice, InvoiceWithLines, PageProps, Verb } from '@/types';
import { List } from './List/Index';
import { AddNewInvoice } from './Shared/AddNewInvoice';
import { useEffect, useState } from 'react';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import Show from './Show';
import { router, usePage } from '@inertiajs/react';
import useCallbackState from '@/hooks/use-callback-state';

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
export default function Index({ auth, invoices, invoice }: PageProps<{ invoices: Invoice[], invoice: InvoiceWithLines }>) {
  const [open, setOpen] = useCallbackState<boolean>(false)
  const page = usePage()
  const hasInvoices = invoices.length > 0;

  const onSelectInvoice = (invoice: Invoice, action: Verb): void => {
    if (action !== "view") return
    setOpen((open) => !open, (newVal) => {
      if (newVal) {
        router.visit(page.url, {
          except: ['invoices'],
          data: {id: invoice.uuid},
          preserveScroll: true,
          preserveState: true
        });
      }
    })
  }

  const onOpenChange = (open: boolean) => {
    setOpen(open);
  };

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

        {invoice && (
          <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-7xl">
            <SheetHeader>
              <SheetTitle>Invoice: {invoice.header.number}</SheetTitle>
              <SheetDescription className="text-[12px]">Invoice details</SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 px-4">
              <Show invoice={invoice} />
            </div>
          </SheetContent>
        </Sheet>
        )}
      </div>
    </AuthenticatedLayout>
  );
}
