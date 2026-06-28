import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import useCallbackState from '@/hooks/use-callback-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { Payable, PageProps, PayableVerb, VendorPaymentWithLines } from '@/types';
import { Link, router, usePage } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { List } from './List/Index';
import { payablesBreadcrumbs } from './constants';
import Show from './Show';

export default function Index({
  auth,
  payables,
  vendorPayment,
  showVendorPayment,
}: PageProps<{ payables: Payable[]; vendorPayment?: VendorPaymentWithLines; showVendorPayment?: boolean }>) {
  const t = useTranslation().trans;
  const [loadingPayment, setLoadingPayment] = useCallbackState<boolean>(false);
  const [open, setOpen] = useCallbackState<boolean>(showVendorPayment ?? false);
  const [selectedPayable, setSelectedPayable] = useCallbackState<Payable | undefined>(undefined);
  const [deleteDialogOpen, setDeleteDialogOpen] = useCallbackState<boolean>(false);
  const page = usePage<PageProps>();
  const hasPayables = payables.length > 0;

  const onSelectPayable = (payable: Payable, action: PayableVerb): void => {
    setSelectedPayable(payable);
    if (action === 'void') {
      setDeleteDialogOpen(true);
      return;
    }
    if (action !== 'view') return;

    setOpen(
      (open) => !open,
      (newVal) => {
        if (newVal) findSelectedPayment(payable.invoice_uuid);
      },
    );
  };

  const findSelectedPayment = (uuid: string) => {
    router.visit(page.url, {
      except: ['payables'],
      data: { id: uuid },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingPayment(true),
      onFinish: () => setLoadingPayment(false),
    });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      router.replace({ url: window.location.pathname, preserveScroll: true, preserveState: true });
    }
  };

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={payablesBreadcrumbs}>
      <div className="space-y-6">
        {hasPayables && (
          <HeadingSmall
            title={t('payables.title')}
            description={t('payables.description')}
            rightPanel={
              <Link href="/payables/create" as="button" className="focus-visible:border-ring focus-visible:ring-ring/50 bg-primary text-primary-foreground hover:bg-primary/90 inline-flex h-9 shrink-0 cursor-pointer items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium whitespace-nowrap shadow-xs transition-all outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 has-[>svg]:px-3 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4">
                <Plus /> {t('payables.newPayment.title')}
              </Link>
            }
          />
        )}

        {!hasPayables && (
          <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
            <h4 className="text-2xl">{t('payables.emptyState.title')}</h4>
            <p className="text-sm text-gray-400">{t('payables.emptyState.description')}</p>
            <Link href="/payables/create" as="button" className="focus-visible:border-ring focus-visible:ring-ring/50 bg-primary text-primary-foreground hover:bg-primary/90 inline-flex h-9 shrink-0 cursor-pointer items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium whitespace-nowrap shadow-xs transition-all outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 has-[>svg]:px-3 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4">
              <Plus /> {t('payables.newPayment.title')}
            </Link>
          </div>
        )}

        {hasPayables && <List data={payables} onSelectPayable={onSelectPayable} />}

        {vendorPayment && !loadingPayment && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-[1380px]">
              <SheetHeader>
                <div className="mr-6 flex items-start justify-between">
                  <div className="flex flex-col">
                    <SheetTitle>{page.props.auth.company.name}</SheetTitle>
                    <SheetDescription className="text-[12px]">{t('payables.viewPayment.description')}</SheetDescription>
                  </div>
                </div>
              </SheetHeader>
              <div className="relative grid gap-4 overflow-y-scroll px-4 pb-4">
                <Show vendorPayment={vendorPayment} auth={auth} />
              </div>
            </SheetContent>
          </Sheet>
        )}

        {selectedPayable && (
          <ConfirmsPassword
            title={t('payables.confirmsPassword.title', { bill: selectedPayable.invoice_number })}
            description={t('payables.confirmsPassword.description')}
            action={t('payables.confirmsPassword.confirm')}
            verb={'update'}
            path={`/payables/${selectedPayable.invoice_uuid}/void`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AppLayout>
  );
}
