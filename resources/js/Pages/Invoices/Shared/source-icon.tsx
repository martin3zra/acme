import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'; // adjust path
import { useTranslation } from '@/hooks/use-translation';
import { InvoiceSource, TransactionKind } from '@/types';
import { Link } from '@inertiajs/react';
import { ExternalLink, FileTextIcon, PackageIcon, RefreshCwIcon } from 'lucide-react';

type SourceIconProp = {
  source: InvoiceSource;
  kind: TransactionKind;
};
export function SourceIcon({ source: { type, id, code }, kind }: SourceIconProp) {
  const t = useTranslation().trans;
  const map: Record<string, { Icon: React.ElementType; tooltip: string }> = {
    template: {
      Icon: RefreshCwIcon,
      tooltip: t(`invoices.source.template`),
    },
    order: {
      Icon: PackageIcon,
      tooltip: t(`invoices.source.order`, { [type]: code as string }),
    },
    estimate: {
      Icon: FileTextIcon,
      tooltip: t(`${kind}s.source.estimate`, { [type]: code as string }),
    },
    invoice: {
      Icon: ExternalLink,
      tooltip: t(`${kind}s.source.${type}`, { [type]: code as string }),
    },
  };

  const entry = map[type];
  if (!entry) return null;

  const { Icon, tooltip } = entry;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        {type === 'template' ? (
          <Icon className="size-4" aria-hidden="true" />
        ) : (
          <Link href={`/${type}s?id=${id}`} title="View Invoice" className="cursor-pointer">
            <Icon className="size-4" aria-hidden="true" />
          </Link>
        )}
      </TooltipTrigger>
      <TooltipContent>{tooltip}</TooltipContent>
    </Tooltip>
  );
}
