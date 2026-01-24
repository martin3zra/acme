import { Button } from '@/components/ui/button';
import { DropdownMenuItem } from '@/components/ui/dropdown-menu';
import { useLocalStorage } from '@/hooks/use-local-storage';
import { useTranslation } from '@/hooks/use-translation';
import { ErrorResponse, InvoiceWithLines, TransactionKind } from '@/types';
import { router } from '@inertiajs/react';
import { ExternalLink, FileDiffIcon } from 'lucide-react';
import { toast } from 'sonner';
import { convertToInvoice } from '../convert-to-invoice';

export function ConvertToInvoiceAction({
  renderedAs,
  kind,
  id,
  source,
}: {
  renderedAs: 'button' | 'dropdown-item';
  kind: TransactionKind;
  source?: InvoiceWithLines;
  id?: string;
}) {
  const t = useTranslation().trans;
  const handleConvertion = () => {
    if (!source) {
      return;
    }
    processConvertion(source);
  };

  const processConvertion = (data: InvoiceWithLines) => {
    const converted = convertToInvoice(kind, data, { terms: 'pia', taxReceipt: 0, ncf: '' });
    const { setItem } = useLocalStorage('invoice');
    setItem(converted);

    router.visit('/invoices/create');
  };

  const handleFetching = async () => {
    if (!id) {
      return;
    }

    const response = await fetch(`/${kind}s/${id}`, {
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
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

  if (source?.header?.source) {
    return (
      <Button onClick={handleRedirection} className="bg-blue-600 hover:bg-blue-700">
        <ExternalLink />
        {t('global.invoiced')}
      </Button>
    );
  }

  if (renderedAs === 'dropdown-item') {
    return <DropdownMenuItem onClick={handleFetching}>{t('global.convertToInvoice')}</DropdownMenuItem>;
  }

  return (
    <Button onClick={handleConvertion} className="bg-green-600 hover:bg-green-700">
      <FileDiffIcon />
      {t('global.convertToInvoice')}
    </Button>
  );
}
