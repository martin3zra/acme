import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import useCallbackState from '@/hooks/use-callback-state';
import { useTranslation } from '@/hooks/use-translation';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { PageProps, Payment, PaymentVerb, PaymentWithLines } from '@/types';
import { Link, router, usePage } from '@inertiajs/react';
import { Ban, NotebookPen, Printer } from 'lucide-react';
import { breadcrumbs } from './constants';
import { List } from './List/Index';
import { AddNewPayment } from './Shared/add-new-payment';
import Show from './Show';

export default function Index({
  auth,
  payments,
  payment,
  showPayment,
}: PageProps<{ payments: Payment[]; payment: PaymentWithLines; showPayment: boolean }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useCallbackState<boolean>(showPayment);
  const [selectedPayment, setSelectedPayment] = useCallbackState<Payment | undefined>(undefined);
  const [deleteDialogOpen, setDeleteDialogOpen] = useCallbackState<boolean>(false);
  const page = usePage<PageProps>();
  const hasPayments = payments.length > 0;

  const onSelectPayment = (payment: Payment, action: PaymentVerb): void => {
    setSelectedPayment(payment);
    if (action === 'void') {
      setDeleteDialogOpen(true);
      return;
    }
    if (action === 'edit') {
      router.visit(`/payments/${payment.uuid}/edit`);
      return;
    }
    if (action !== 'view') return;

    setOpen(
      (open) => !open,
      (newVal) => {
        if (newVal) findSelectedPayment(payment.uuid);
      },
    );
  };

  const findSelectedPayment = (uuid: string) => {
    router.visit(page.url, {
      except: ['payments'],
      data: { id: uuid },
      preserveScroll: true,
      preserveState: true,
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

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };
  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasPayments && <HeadingSmall title={t('payments.title')} description={t('payments.description')} rightPanel={<AddNewPayment />} />}

        {!hasPayments && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t('payments.emptyState.title')}</h4>
              <p className="text-sm text-gray-400">{t('payments.emptyState.description')}</p>
              <AddNewPayment />
            </div>
          </>
        )}

        {hasPayments && <List data={payments} onSelectPayment={onSelectPayment} />}

        {payment && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-[1380px]">
              <SheetHeader>
                <div className="mr-6 flex items-start justify-between">
                  <div className="flex flex-col">
                    <SheetTitle>{page.props.auth.company.name}</SheetTitle>
                    <SheetDescription className="text-[12px]">{t('payments.viewPayment.description')}</SheetDescription>
                  </div>
                  <div className="mx-4 flex gap-x-3">
                    {payment.header.status !== 'void' && (
                      <>
                        <Button variant={'destructive'} onClick={() => onSelectPayment(payment.header, 'void')}>
                          <Ban /> {t('global.actions.void')}
                        </Button>
                        <Separator orientation="vertical" />
                        <Button asChild disabled={payment.header.status === 'void'}>
                          <Link href={`/payments/${payment.header.uuid}/edit`} as="button">
                            <NotebookPen /> {t('global.actions.edit')}
                          </Link>
                        </Button>
                      </>
                    )}

                    <Button>
                      <Printer /> {t('global.actions.print')}
                    </Button>
                  </div>
                </div>
              </SheetHeader>
              <div className="relative grid gap-4 px-4">
                {payment.header.status === 'void' && (
                  <div className="absolute inset-0 flex w-full items-center justify-center overflow-y-hidden bg-transparent">
                    <h1 className="-rotate-45 border-8 border-red-500/25 p-8 text-8xl font-extrabold text-red-500/25">VOID</h1>
                  </div>
                )}
                <Show payment={payment} auth={auth} />
              </div>
            </SheetContent>
          </Sheet>
        )}

        {selectedPayment && (
          <ConfirmsPassword
            title={t('payments.confirmsPassword.title', { payment: selectedPayment.number })}
            description={t('payments.confirmsPassword.description', { total: selectedPayment.amount })}
            action={t('payments.confirmsPassword.confirm')}
            verb={'update'}
            path={`/payments/${selectedPayment.uuid}/void`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AuthenticatedLayout>
  );
}
