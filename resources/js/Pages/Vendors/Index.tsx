import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { ImportDrawer } from '@/components/import-zone';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import useCallbackState from '@/hooks/use-callback-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps, Payable, TaxReceipt, Vendor, VendorTypeFilter, VendorVerb } from '@/types';
import { router, usePage } from '@inertiajs/react';
import { FileUp, Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { List } from './List/Index';
import CreateForm, { CreateFormParams } from './Shared/CreateForm';
import { breadcrumbs } from './constants';

export default function Index({
  auth,
  vendors,
  vendor,
  tax_receipts,
  currentVendorTypeFilter,
  openState,
  payables,
}: PageProps<{
  openState: boolean;
  vendors: Vendor[];
  vendor: Vendor;
  tax_receipts: TaxReceipt[];
  currentVendorTypeFilter: VendorTypeFilter;
  payables?: Payable[];
}>) {
  const t = useTranslation().trans;
  const page = usePage<PageProps>();
  const [loadingVendor, setLoadingVendor] = useState<boolean>(false);
  const [open, setOpen] = useCallbackState<boolean>(vendor !== undefined || openState);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [importSheetOpen, setImportSheetOpen] = useState<boolean>(false);
  const [selectedVendor, setSelectedVendor] = useState<CreateFormParams>({
    vendor: undefined,
    action: vendor !== undefined ? 'view' : 'create',
    tax_receipts: tax_receipts,
  });

  useEffect(() => {
    if (vendor === undefined) return;
    setSelectedVendor((val) => ({ ...val, vendor }));
  }, [vendor, setSelectedVendor]);

  const verbName = useVerb().action(selectedVendor.action);
  const hasVendors = vendors.length > 0;

  const onCreateNewVendor = () => {
    setSelectedVendor({ vendor: undefined, action: 'create', tax_receipts });
    setOpen(true);
  };

  const onSelectVendor = (vendor: Vendor, action: VendorVerb): void => {
    if (action === 'record-payment' || action === 'record-purchase') {
      const url = action === 'record-payment' ? '/payments/create' : '/purchases/create';
      router.visit(url, { data: { vendor_id: vendor.uuid } });
      return;
    }

    setSelectedVendor({ vendor, action, tax_receipts });

    if (action === 'trash') {
      setDeleteDialogOpen(true);
      return;
    }

    if (open) return;
    setOpen(
      (open) => !open,
      (newVal) => {
        if (newVal) findSelectedVendor(vendor.uuid);
      },
    );
  };

  const findSelectedVendor = (uuid: string) => {
    router.visit(page.url, {
      except: ['vendors', 'tax_receipts'],
      data: { id: uuid },
      preserveScroll: true,
      preserveState: true,
    });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      setSelectedVendor({ vendor: undefined, action: 'create', tax_receipts });
      // Remove query string from URL
      router.replace({
        url: window.location.pathname,
        preserveScroll: true,
        preserveState: true,
      });
    }
  };

  useEffect(() => {
    if (selectedVendor && selectedVendor.vendor !== undefined) {
      if (selectedVendor.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedVendor, setOpen]);

  const onVendorFilterTypeChange = (value: VendorTypeFilter) => {
    router.visit(page.url, {
      data: { vendorType: value },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingVendor(true),
      onFinish: () => setLoadingVendor(false),
    });
  };

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasVendors && (
          <HeadingSmall
            title={t('vendors.title')}
            description={t('vendors.description')}
            rightPanel={
              <div className="flex space-x-2">
                <Button onClick={onCreateNewVendor}>
                  <Plus /> {t('vendors.newVendor.title')}
                </Button>
                <Button onClick={() => setImportSheetOpen(true)}>
                  <FileUp /> {t('global.actions.import')}
                </Button>
              </div>
            }
          />
        )}

        {!hasVendors && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-61 min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-3xl bg-white p-10 shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t(`vendors.emptyState.${currentVendorTypeFilter}.title`)}</h4>
              <p className="text-sm text-gray-400">{t(`vendors.emptyState.${currentVendorTypeFilter}.description`)}</p>
              <div className="flex space-x-3">
                {currentVendorTypeFilter !== 'all' && (
                  <Button variant={'outline'} onClick={() => onVendorFilterTypeChange('all')}>
                    {t('vendors.viewAll')}
                  </Button>
                )}

                <Button onClick={onCreateNewVendor}>
                  <Plus /> {t('vendors.newVendor.title')}
                </Button>

                <Button onClick={() => setImportSheetOpen(true)}>
                  <FileUp /> {t('global.actions.import')}
                </Button>
              </div>
            </div>
          </>
        )}

        {hasVendors && (
          <List
            data={vendors}
            currentVendorTypeFilter={currentVendorTypeFilter}
            onSelectVendor={onSelectVendor}
            onVendorTypeFilterChanges={onVendorFilterTypeChange}
          />
        )}

        {!loadingVendor && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-7xl">
              <SheetHeader>
                <SheetTitle>
                  {t(`global.actions.${verbName}`)} {t(`global.vendor`).toLocaleLowerCase()}
                </SheetTitle>
                <SheetDescription className="text-[12px]">{t(`vendors.newVendor.description`)}</SheetDescription>
              </SheetHeader>
              <div className="no-scrollbar grid gap-4 overflow-y-scroll px-4">
                <CreateForm params={selectedVendor} onFinish={() => modalHandler(false)} />
                {selectedVendor.action === 'view' && payables && payables.length > 0 && (
                  <div className="mt-4 px-4">
                    <div className="mb-2 flex items-center justify-between">
                      <h3 className="text-sm font-semibold">{t('payables.title')}</h3>
                      <a
                        href={`/payables/create?vendor_id=${selectedVendor.vendor?.uuid}`}
                        className="text-primary text-xs hover:underline"
                      >
                        {t('payables.recordPayment')}
                      </a>
                    </div>
                    <table className="w-full text-sm [&_td]:p-1.5 [&_th]:p-1.5 [&_th]:text-left [&_th]:font-medium [&_th]:text-gray-500">
                      <thead>
                        <tr className="border-b">
                          <th>{t('global.number')}</th>
                          <th>{t('global.dueDate')}</th>
                          <th className="text-right">{t('global.balance')}</th>
                          <th>{t('global.status')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {payables.map((p) => (
                          <tr key={p.uuid} className="border-b last:border-0">
                            <td>{p.invoice_number}</td>
                            <td>{p.due_date}</td>
                            <td className="text-right">{(p.amount_payable - p.amount_paid).toFixed(2)}</td>
                            <td>
                              <span className={`rounded px-1.5 py-0.5 text-xs font-medium ${p.paid_status === 'paid' ? 'bg-green-100 text-green-700' : new Date(p.due_date) < new Date() ? 'bg-red-100 text-red-700' : 'bg-yellow-100 text-yellow-700'}`}>
                                {p.paid_status}
                              </span>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </SheetContent>
          </Sheet>
        )}
        {selectedVendor.vendor && (
          <ConfirmsPassword
            title={t(`vendors.confirmsPassword.title`, { vendor: selectedVendor?.vendor?.name })}
            description={t(`vendors.confirmsPassword.description`, { vendor: selectedVendor?.vendor?.name })}
            action={t(`vendors.confirmsPassword.confirm`)}
            verb={'destroy'}
            path={`/vendors/${selectedVendor?.vendor?.id}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
        <ImportDrawer source="vendors" openImportDrawer={importSheetOpen} setImportDrawer={setImportSheetOpen} />
      </div>
    </AppLayout>
  );
}
