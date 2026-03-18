import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useTranslation } from '@/hooks/use-translation';
import { PurchaseSource, PurchaseTransactionKind } from '@/types';
import { Link } from '@inertiajs/react';
import { ExternalLink, PackageIcon } from 'lucide-react';

type PurchaseSourceIconProp = {
  source: PurchaseSource;
  kind: PurchaseTransactionKind;
};

const purchaseSourceUrl: Record<string, string> = {
  purchase_order: '/purchases/orders',
  purchase_receipt: '/purchases/receipts',
  vendor_bill: '/purchases/vendor-bills',
};

export function PurchaseSourceIcon({ source: { type, id, code }, kind }: PurchaseSourceIconProp) {
  const t = useTranslation().trans;
  const map: Record<string, { Icon: React.ElementType; tooltip: string }> = {
    purchase_order: {
      Icon: PackageIcon,
      tooltip: t('purchases.source.purchase_order', { purchase_order: code as string }),
    },
    purchase_receipt: {
      Icon: ExternalLink,
      tooltip: t('purchases.source.purchase_receipt', { purchase_receipt: code as string }),
    },
  };

  const entry = map[type];
  if (!entry) return null;

  const { Icon, tooltip } = entry;
  const base = purchaseSourceUrl[type] ?? '/purchases/orders';

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link href={`${base}?id=${id}`} className="cursor-pointer">
          <Icon className="size-4" aria-hidden="true" />
        </Link>
      </TooltipTrigger>
      <TooltipContent>{tooltip}</TooltipContent>
    </Tooltip>
  );
}
