import { ConfirmsPassword } from '@/components/confirms-password';
import { Button } from '@/components/ui/button';
import { useTranslation } from '@/hooks/use-translation';
import type { PurchaseTransactionKind, PurchaseWithLines } from '@/types';
import { CheckCircle } from 'lucide-react';
import { useState } from 'react';

type Props = {
  purchase: PurchaseWithLines;
  kind: PurchaseTransactionKind;
};

export function ConfirmPurchaseAction({ purchase, kind }: Props) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);

  const isReceipt = kind === 'purchase_receipt';
  const label = isReceipt ? t('purchases.receipts.confirm') : t('purchases.vendorBills.post');
  const description = isReceipt ? t('purchases.receipts.confirmDescription') : t('purchases.vendorBills.postDescription');

  if (purchase.header.status !== 'draft') return null;

  return (
    <>
      <Button variant="default" className="w-full" onClick={() => setOpen(true)}>
        <CheckCircle className="mr-2 size-4" />
        {label}
      </Button>

      <ConfirmsPassword
        open={open}
        onOpenChange={setOpen}
        title={label}
        description={description}
        action={label}
        verb="update"
        path={`/purchases/${purchase.header.uuid}/confirm`}
      />
    </>
  );
}
