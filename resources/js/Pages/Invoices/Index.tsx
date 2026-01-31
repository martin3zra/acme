import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import useCallbackState from '@/hooks/use-callback-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { capitalize } from '@/lib/utils';
import { Invoice, InvoiceTypeFilter, InvoiceVerb, InvoiceWithLines, PageProps, TransactionKind } from '@/types';
import { Link, router, usePage } from '@inertiajs/react';
import { Ban, DollarSign, NotebookPen, Printer } from 'lucide-react';
import { makeBreadcrumbs } from './constants';
import { List } from './List/Index';
import { AddNewInvoice } from './Shared/AddNewInvoice';
import { ConvertToInvoiceAction } from './Shared/convert-to-invoice-action';
import Show from './Show';

export default function Index({
  auth,
  invoices,
  invoice,
  showInvoice,
  currentInvoiceTypeFilter,
  kind,
}: PageProps<{
  invoices: Invoice[];
  invoice: InvoiceWithLines;
  showInvoice: boolean;
  currentInvoiceTypeFilter: InvoiceTypeFilter;
  kind: TransactionKind;
}>) {
  const isInvoice = kind === 'invoice';
  const [loadingInvoice, setLoadingInvoice] = useCallbackState<boolean>(false);
  const [open, setOpen] = useCallbackState<boolean>(showInvoice);
  const [selectedInvoice, setSelectedInvoice] = useCallbackState<Invoice | undefined>(undefined);
  const [deleteDialogOpen, setDeleteDialogOpen] = useCallbackState<boolean>(false);
  const page = usePage<PageProps>();
  const hasInvoices = invoices.length > 0;
  const t = useTranslation().trans;

  const screen = {
    invoice: { key: 'invoices', title: t('invoices.newInvoice.title'), url: '/invoices/create' },
    template: { key: 'invoices', title: t('invoices.newInvoice.title'), url: '/invoices/create' },
    estimate: { key: 'estimates', title: t('estimates.newEstimate.title'), url: '/estimates/create' },
    order: { key: 'orders', title: t('orders.newOrder.title'), url: '/orders/create' },
  }[kind];

  const onSelectInvoice = (invoice: Invoice, action: InvoiceVerb): void => {
    setSelectedInvoice(invoice);
    if (isInvoice) {
      if (action === 'record-payment') {
        router.visit(`/payments/create`, { data: { customer_id: invoice.customer.uuid, invoice_id: invoice.uuid } });
        return;
      }
      if (action === 'void') {
        setDeleteDialogOpen(true);
        return;
      }
    }
    if (action === 'edit') {
      router.visit(`/${kind}s/${invoice.uuid}/edit`);
      return;
    }
    if (action !== 'view') return;

    if (open) return;
    setOpen(
      (open) => !open,
      (newVal) => {
        if (newVal) findSelectedInvoice(invoice.uuid);
      },
    );
  };

  const findSelectedInvoice = (uuid: string) => {
    router.visit(page.url, {
      except: ['invoices'],
      data: { id: uuid },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingInvoice(true),
      onFinish: () => setLoadingInvoice(false),
    });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      // Remove query string from URL
      router.replace({
        url: window.location.pathname,
        preserveScroll: true,
        preserveState: true,
      });
    }
  };

  const onInvoiceFilterTypeChange = (value: InvoiceTypeFilter) => {
    router.visit(page.url, {
      data: { invoiceType: value },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingInvoice(true),
      onFinish: () => setLoadingInvoice(false),
    });
  };

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={makeBreadcrumbs(kind)}>
      <div className="space-y-6">
        {hasInvoices && (
          <HeadingSmall
            title={t(`${screen.key}.title`)}
            description={t(`${screen.key}.description`)}
            rightPanel={<AddNewInvoice title={screen.title} url={screen.url} />}
          />
        )}

        {!hasInvoices && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t(`${screen.key}.emptyState.title`)}</h4>
              <p className="text-sm text-gray-400">{t(`${screen.key}.emptyState.description`)}</p>
              <AddNewInvoice title={screen.title} url={screen.url} />
            </div>
          </>
        )}

        {hasInvoices && (
          <List
            kind={kind}
            data={invoices}
            currentInvoiceTypeFilter={currentInvoiceTypeFilter}
            onSelectInvoice={onSelectInvoice}
            onInvoiceTypeFilterChanges={onInvoiceFilterTypeChange}
          />
        )}

        {invoice && !loadingInvoice && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-[1380px]">
              <SheetHeader>
                <div className="mr-6 flex items-start justify-between">
                  <div className="flex flex-col">
                    <SheetTitle>{page.props.auth.company.name}</SheetTitle>
                    <SheetDescription className="text-[12px]">{t(`${kind}s.view${capitalize(kind)}.description`)}</SheetDescription>
                  </div>
                  <div className="mx-4 flex gap-x-3">
                    {invoice.header.status !== 'void' && invoice.header.status !== 'closed' && (
                      <>
                        {isInvoice && (
                          <Button variant={'destructive'} onClick={() => onSelectInvoice(invoice.header, 'void')}>
                            <Ban /> {t('global.actions.void')}
                          </Button>
                        )}
                        <Separator orientation="vertical" />
                        <Button asChild disabled={invoice.header.status === 'void'}>
                          <Link href={`/${kind}s/${invoice.header.uuid}/edit`} as="button">
                            <NotebookPen /> {t('global.actions.edit')}
                          </Link>
                        </Button>
                        {isInvoice && (invoice.header.paid_status === 'unpaid' || invoice.header.paid_status === 'partial') && (
                          <Button asChild disabled={invoice.header.status === 'void'}>
                            <Link href={`/payments/create?customer_id=${invoice.header.customer.uuid}&invoice_id=${invoice.header.uuid}`} as="button">
                              <DollarSign /> {t('global.actions.recordPayment')}
                            </Link>
                          </Button>
                        )}
                      </>
                    )}

                    {kind === 'estimate' && (
                      <ConvertToInvoiceAction title={t('global.convertToInvoice')} renderedAs="button" kind={kind} source={invoice} />
                    )}
                    {kind === 'invoice' && (
                      <ConvertToInvoiceAction
                        mode="duplicate"
                        title={t('global.duplicateInvoice')}
                        renderedAs="button"
                        kind={kind}
                        source={invoice}
                      />
                    )}
                    <a
                      href={invoice.pdfURL}
                      className="flex items-center gap-x-3 rounded-sm bg-indigo-600 px-4 text-white hover:bg-indigo-700"
                      target="_blank"
                    >
                      <Printer className="size-4" /> {t('global.actions.print')}
                    </a>
                  </div>
                </div>
              </SheetHeader>
              <div className="relative grid gap-4 overflow-y-scroll px-4 pb-4">
                {invoice.header.status === 'void' && (
                  <div className="absolute inset-0 flex w-full items-center justify-center overflow-y-hidden bg-transparent">
                    <h1 className="-rotate-45 border-8 border-red-500/25 p-8 text-8xl font-extrabold text-red-500/25 uppercase">
                      {t('global.voided')}
                    </h1>
                  </div>
                )}
                <Show kind={kind} invoice={invoice} auth={auth} />
              </div>
            </SheetContent>
          </Sheet>
        )}

        {selectedInvoice && (
          <ConfirmsPassword
            title={t('invoices.confirmsPassword.title', { invoice: selectedInvoice.number })}
            description={t('invoices.confirmsPassword.description', { total: selectedInvoice.total })}
            action={t('invoices.confirmsPassword.confirm')}
            verb={'update'}
            path={`/invoices/${selectedInvoice.uuid}/void`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AppLayout>
  );
}
