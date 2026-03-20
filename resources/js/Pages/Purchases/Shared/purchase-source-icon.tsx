import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useTranslation } from '@/hooks/use-translation';
import { PurchaseSource, PurchaseTransactionKind } from '@/types';
import { Link } from '@inertiajs/react';
import { ExternalLink, FileText, PackageIcon } from 'lucide-react';

type PurchaseSourceIconProp = {
  source: PurchaseSource;
  kind: PurchaseTransactionKind;
};

const purchaseSourceUrl: Record<string, string> = {
  purchase_order: '/purchases/orders',
  purchase_receipt: '/purchases/receipts',
  vendor_bill: '/purchases/vendor-bills',
};

type SourceEntry = { Icon: React.ElementType; tooltip: string; id: string | number; type: string };

function SourceIconLink({ Icon, tooltip, href }: { Icon: React.ElementType; tooltip: string; href: string }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link href={href} className="cursor-pointer">
          <Icon className="size-4" aria-hidden="true" />
        </Link>
      </TooltipTrigger>
      <TooltipContent>{tooltip}</TooltipContent>
    </Tooltip>
  );
}

export function PurchaseSourceIcon({ source, kind }: PurchaseSourceIconProp) {
  const t = useTranslation().trans;

  const backLinkMap: Record<string, { Icon: React.ElementType; tooltip: string }> = {
    purchase_order: {
      Icon: PackageIcon,
      tooltip: t('purchases.source.purchase_order', { purchase_order: source.code as string }),
    },
    purchase_receipt: {
      Icon: ExternalLink,
      tooltip: t('purchases.source.purchase_receipt', { purchase_receipt: source.code as string }),
    },
  };

  const backEntry = backLinkMap[source.type];
  if (!backEntry) return null;

  const backBase = purchaseSourceUrl[source.type] ?? '/purchases/orders';
  const entries: SourceEntry[] = [{ ...backEntry, id: source.id, type: source.type }];

  if (source.target) {
    entries.push({
      Icon: FileText,
      tooltip: t('purchases.source.to_vendor_bill', { vendor_bill: source.target.code as string }),
      id: source.target.id,
      type: source.target.type,
    });
  }

  return (
    <div className="flex items-center gap-1">
      {entries.map((entry) => {
        const base = purchaseSourceUrl[entry.type] ?? '/purchases/orders';
        return (
          <SourceIconLink
            key={entry.type}
            Icon={entry.Icon}
            tooltip={entry.tooltip}
            href={`${base}?id=${entry.id}`}
          />
        );
      })}
    </div>
  );
}
