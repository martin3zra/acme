import { Button } from '@/components/ui/button';
import { DropdownMenuItem } from '@/components/ui/dropdown-menu';
import { useLocalStorage } from '@/hooks/use-local-storage';
import type { ErrorResponse, PurchaseWithLines } from '@/types';
import { router } from '@inertiajs/react';
import { FileDiffIcon } from 'lucide-react';
import { toast } from 'sonner';
import { convertToReceipt } from '../convert-to-receipt';

type Props = {
  title: string;
  renderedAs?: 'button' | 'dropdown-item';
  id?: string;
  source?: PurchaseWithLines;
};

export function ConvertToReceiptAction({ title, renderedAs = 'dropdown-item', id, source }: Props) {
  const { setItem } = useLocalStorage('purchase_purchase_receipt');

  const processConversion = (data: PurchaseWithLines) => {
    const converted = convertToReceipt(data);
    setItem(converted);
    router.visit('/purchases/receipts/create');
  };

  const handleFetching = async () => {
    if (!id) return;

    const response = await fetch(`/purchases/${id}`, {
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      toast.error(error.status);
      throw new Error('Failed fetching the data');
    }

    const data: PurchaseWithLines = await response.json();
    processConversion(data);
  };

  const handleConversion = () => {
    if (source) {
      processConversion(source);
      return;
    }

    if (id) {
      void handleFetching();
    }
  };

  if (renderedAs === 'dropdown-item') {
    return <DropdownMenuItem onClick={handleFetching}>{title}</DropdownMenuItem>;
  }

  return (
    <Button onClick={handleConversion} className="bg-green-600 hover:bg-green-700">
      <FileDiffIcon />
      {title}
    </Button>
  );
}
