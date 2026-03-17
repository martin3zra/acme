import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import type { PageProps, Purchase, PurchaseTransactionKind, PurchaseWithLines } from '@/types';
import { Link, router, usePage } from '@inertiajs/react';
import { NotebookPen, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { makeBreadcrumbs, purchaseKindMeta } from './constants';
import { List } from './List/Index';
import { AddNewPurchase } from './Shared/AddNewPurchase';
import { ConvertToReceiptAction } from './Shared/convert-to-receipt-action';
import Show from './Show';

export default function Index({
  auth,
  purchases,
  purchase,
  showPurchase,
  kind,
}: PageProps<{
  purchases: Purchase[];
  purchase: PurchaseWithLines;
  showPurchase: boolean;
  kind: PurchaseTransactionKind;
}>) {
  const t = useTranslation().trans;
  const page = usePage<PageProps>();
  const meta = purchaseKindMeta(kind);

  const [open, setOpen] = useState<boolean>(showPurchase);
  const [loadingPurchase, setLoadingPurchase] = useState<boolean>(false);
  const [selectedPurchase, setSelectedPurchase] = useState<Purchase | undefined>(undefined);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);

  const hasPurchases = purchases.length > 0;

  const onSelectPurchase = (purchase: Purchase, action: 'view' | 'edit' | 'delete'): void => {
    setSelectedPurchase(purchase);

    if (action === 'delete') {
      setDeleteDialogOpen(true);
      return;
    }

    if (action === 'edit') {
      router.visit(`/purchases/${purchase.uuid}/edit`);
      return;
    }

    if (action !== 'view') return;

    if (open) return;
    setOpen(true);
    findSelectedPurchase(purchase.uuid);
  };

  const findSelectedPurchase = (uuid: string) => {
    router.visit(page.url, {
      except: ['purchases'],
      data: { id: uuid },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingPurchase(true),
      onFinish: () => setLoadingPurchase(false),
    });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      router.replace({
        url: window.location.pathname,
        preserveScroll: true,
        preserveState: true,
      });
    }
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={makeBreadcrumbs(kind)}>
      <div className="space-y-6">
        {hasPurchases && (
          <HeadingSmall
            title={t(`${meta.key}.title`)}
            description={t(`${meta.key}.description`)}
            rightPanel={<AddNewPurchase title={t(`${meta.key}.new.title`)} url={meta.createUrl} />}
          />
        )}

        {!hasPurchases && (
          <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
            <h4 className="text-2xl">{t(`${meta.key}.emptyState.title`)}</h4>
            <p className="text-sm text-gray-400">{t(`${meta.key}.emptyState.description`)}</p>
            <AddNewPurchase title={t(`${meta.key}.new.title`)} url={meta.createUrl} />
          </div>
        )}

        {hasPurchases && <List data={purchases} onSelectPurchase={onSelectPurchase} />}

        {purchase && !loadingPurchase && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-[1380px]">
              <SheetHeader>
                <div className="mr-6 flex items-start justify-between">
                  <div className="flex flex-col">
                    <SheetTitle>{page.props.auth.company.name}</SheetTitle>
                    <SheetDescription className="text-[12px]">{t(`${meta.key}.view.description`)}</SheetDescription>
                  </div>
                  <div className="mx-4 flex gap-x-3">
                    <Button variant={'destructive'} onClick={() => setDeleteDialogOpen(true)}>
                      <Trash2 /> {t('global.actions.delete')}
                    </Button>
                    <Separator orientation="vertical" />
                    <Button asChild>
                      <Link href={`/purchases/${purchase.header.uuid}/edit`} as="button">
                        <NotebookPen /> {t('global.actions.edit')}
                      </Link>
                    </Button>

                    {kind === 'purchase_order' && (
                      <ConvertToReceiptAction
                        title={t('global.convertToReceipt')}
                        renderedAs="button"
                        source={purchase}
                      />
                    )}
                  </div>
                </div>
              </SheetHeader>
              <div className="relative grid gap-4 overflow-y-scroll px-4 pb-4">
                <Show kind={kind} purchase={purchase} auth={auth} />
              </div>
            </SheetContent>
          </Sheet>
        )}

        {selectedPurchase && (
          <ConfirmsPassword
            open={deleteDialogOpen}
            onOpenChange={setDeleteDialogOpen}
            verb="destroy"
            path={`/purchases/${selectedPurchase.uuid}`}
            title={t('purchases.deletePurchase.title')}
            description={t('purchases.deletePurchase.description')}
            action={t('global.actions.delete')}
          />
        )}
      </div>
    </AppLayout>
  );
}
