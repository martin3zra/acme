import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { ImportDrawer } from '@/components/import-zone';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import useCallbackState from '@/hooks/use-callback-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { Customer, CustomerTypeFilter, CustomerVerb, PageProps, TaxReceipt } from '@/types';
import { router, usePage } from '@inertiajs/react';
import { FileUp, Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { List } from './List/Index';
import CreateForm, { CreateFormParams } from './Shared/CreateForm';
import { breadcrumbs } from './constants';

export default function Index({
  auth,
  customers,
  customer,
  tax_receipts,
  currentCustomerTypeFilter,
}: PageProps<{ customers: Customer[]; customer: Customer; tax_receipts: TaxReceipt[]; currentCustomerTypeFilter: CustomerTypeFilter }>) {
  const t = useTranslation().trans;
  const page = usePage<PageProps>();
  const [loadingCustomer, setLoadingCustomer] = useState<boolean>(false);
  const [open, setOpen] = useCallbackState<boolean>(customer !== undefined);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [importSheetOpen, setImportSheetOpen] = useState<boolean>(false);
  const [selectedCustomer, setSelectedCustomer] = useState<CreateFormParams>({
    customer: undefined,
    action: customer !== undefined ? 'view' : 'create',
    tax_receipts: tax_receipts,
  });

  useEffect(() => {
    if (customer === undefined) return;
    setSelectedCustomer((val) => ({ ...val, customer }));
  }, [customer, setSelectedCustomer]);

  const verbName = useVerb().action(selectedCustomer.action);
  const hasCustomers = customers.length > 0;

  const onCreateNewCustomer = () => {
    setSelectedCustomer({ customer: undefined, action: 'create', tax_receipts });
    setOpen(true);
  };

  const onSelectCustomer = (customer: Customer, action: CustomerVerb): void => {
    if (action === 'record-payment' || action === 'issue-invoice') {
      const url = action === 'record-payment' ? '/payments/create' : '/invoices/create';
      router.visit(url, { data: { customer_id: customer.uuid } });
      return;
    }

    setSelectedCustomer({ customer, action, tax_receipts });

    if (open) return;
    setOpen(
      (open) => !open,
      (newVal) => {
        if (newVal) findSelectedCustomer(customer.uuid);
      },
    );
  };

  const findSelectedCustomer = (uuid: string) => {
    router.visit(page.url, {
      except: ['customers', 'tax_receipts'],
      data: { id: uuid },
      preserveScroll: true,
      preserveState: true,
    });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      setSelectedCustomer({ customer: undefined, action: 'create', tax_receipts });
      // Remove query string from URL
      router.replace({
        url: window.location.pathname,
        preserveScroll: true,
        preserveState: true,
      });
    }
  };

  useEffect(() => {
    if (selectedCustomer && selectedCustomer.customer !== undefined) {
      if (selectedCustomer.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedCustomer, setOpen]);

  const onCustomerFilterTypeChange = (value: CustomerTypeFilter) => {
    router.visit(page.url, {
      data: { customerType: value },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingCustomer(true),
      onFinish: () => setLoadingCustomer(false),
    });
  };

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasCustomers && (
          <HeadingSmall
            title={t('customers.title')}
            description={t('customers.description')}
            rightPanel={
              <div className="flex space-x-2">
                <Button onClick={onCreateNewCustomer}>
                  <Plus /> {t('customers.newCustomer.title')}
                </Button>
                <Button onClick={() => setImportSheetOpen(true)}>
                  <FileUp /> {t('global.actions.import')}
                </Button>
              </div>
            }
          />
        )}

        {!hasCustomers && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t('customers.emptyState.title')}</h4>
              <p className="text-sm text-gray-400">{t('customers.emptyState.description')}</p>
              <Button onClick={onCreateNewCustomer}>
                <Plus /> {t('customers.newCustomer.title')}
              </Button>
            </div>
          </>
        )}

        {hasCustomers && (
          <List
            data={customers}
            currentCustomerTypeFilter={currentCustomerTypeFilter}
            onSelectCustomer={onSelectCustomer}
            onCustomerTypeFilterChanges={onCustomerFilterTypeChange}
          />
        )}

        {!loadingCustomer && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-7xl">
              <SheetHeader>
                <SheetTitle>
                  {t(`global.actions.${verbName}`)} {t(`global.customer`).toLocaleLowerCase()}
                </SheetTitle>
                <SheetDescription className="text-[12px]">{t(`customers.newCustomer.description`)}</SheetDescription>
              </SheetHeader>
              <div className="no-scrollbar grid gap-4 overflow-y-scroll px-4">
                <CreateForm params={selectedCustomer} onFinish={() => modalHandler(false)} />
              </div>
            </SheetContent>
          </Sheet>
        )}
        {selectedCustomer.customer && (
          <ConfirmsPassword
            title={t(`customers.confirmsPassword.title`, { customer: selectedCustomer?.customer?.name })}
            description={t(`customers.confirmsPassword.description`, { customer: selectedCustomer?.customer?.name })}
            action={t(`customers.confirmsPassword.confirm`)}
            verb={'destroy'}
            path={`/customers/${selectedCustomer?.customer?.id}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
        <ImportDrawer source="customers" openImportDrawer={importSheetOpen} setImportDrawer={setImportSheetOpen} />
      </div>
    </AppLayout>
  );
}
