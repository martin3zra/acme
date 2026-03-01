import { Button } from '@/components/ui/button';
import { DropdownMenuItem } from '@/components/ui/dropdown-menu';
import { useLocalStorage } from '@/hooks/use-local-storage';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { ErrorResponse, InvoiceWithLines, TransactionKind } from '@/types';
import { router } from '@inertiajs/react';
import { Copy, ExternalLink, FileDiffIcon } from 'lucide-react';
import { toast } from 'sonner';
import { convertToInvoice } from '../convert-to-invoice';

type Props = {
  mode?: 'convert' | 'duplicate';
  title: string;
  renderedAs?: 'button' | 'dropdown-item';
  kind: TransactionKind;
  source?: InvoiceWithLines;
  id?: string;
};

export function ConvertToInvoiceAction({ mode = 'convert', title, renderedAs = 'dropdown-item', kind, id, source }: Props) {
  const t = useTranslation().trans;
  const handleConvertion = () => {
    if (!source) {
      return;
    }
    processConvertion(source);
  };

  const processConvertion = (data: InvoiceWithLines) => {
    // const terms: PaymentTermValue = mode === 'convert' ? 'pia' : data.header.terms;
    const clonedFrom: number | undefined = mode === 'duplicate' ? data.header.id : undefined;
    const taxReceipt: number = mode === 'duplicate' ? data.header.tax_receipt_id : 0;
    const converted = convertToInvoice(kind, data, { terms: data.header.terms, taxReceipt, ncf: '' }, clonedFrom);
    const { setItem } = useLocalStorage('invoice');
    setItem(converted);

    router.visit('/invoices/create');
  };

  const handleFetching = async () => {
    if (!id) {
      return;
    }

    const response = await fetch(`/invoices/${id}`, {
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        'X-Transaction-Kind': kind,
      },
      credentials: 'include',
    });
    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      toast.error(error.status);
      throw new Error('Failed fetching the data', error.data);
    }
    const data: InvoiceWithLines = await response.json();
    processConvertion(data);
  };

  const handleRedirection = () => {
    router.visit(`/invoices?id=${source?.header?.source?.id}`);
  };

  if (source?.header?.source && mode === 'convert') {
    return (
      <Button onClick={handleRedirection} className="bg-blue-600 hover:bg-blue-700">
        <ExternalLink />
        {t('global.invoiced')}
      </Button>
    );
  }

  if (renderedAs === 'dropdown-item') {
    return <DropdownMenuItem onClick={handleFetching}>{title}</DropdownMenuItem>;
  }

  return (
    <Button
      onClick={handleConvertion}
      className={cn([
        mode === 'convert' && 'bg-green-600 hover:bg-green-700',
        mode === 'duplicate' && 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100',
      ])}
    >
      {mode === 'convert' ? <FileDiffIcon /> : <Copy />}
      {title}
    </Button>
  );
}
