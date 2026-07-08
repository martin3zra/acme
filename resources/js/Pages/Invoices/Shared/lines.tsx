import { AlertDestructive } from '@/components/alert-destructive';
import { Command, CommandDialog, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { useTranslation } from '@/hooks/use-translation';
import { Item, LineForm, TransactionKind } from '@/types';
import { router } from '@inertiajs/react';
import { LayoutListIcon } from 'lucide-react';
import { useEffect, useState } from 'react';
import LinesColumnHeaders from './lines-column-headers';
import LinesForm from './lines-form';
import Line from './lines-line';

type LinesProps = {
  kind: TransactionKind;
  items: Item[];
  lines: LineForm[];
  lineError?: string;
  currentItem: Item | undefined;
  handleRemoveLine: (event: React.MouseEvent<HTMLButtonElement>) => void;
  handleKeyDown: (event: React.KeyboardEvent<HTMLInputElement>) => void;
  handleOnSelected: (item: Item) => void;
  amount: number;
  setAmount: React.Dispatch<React.SetStateAction<number>>;
  referenceInputRef: React.RefObject<HTMLInputElement | null>;
  qtyInputRef: React.RefObject<HTMLInputElement | null>;
};
export const Lines = ({
  kind,
  lineError,
  referenceInputRef,
  qtyInputRef,
  lines,
  currentItem,
  handleRemoveLine,
  handleKeyDown,
  handleOnSelected,
  amount,
  setAmount,
  items,
}: LinesProps) => {
  const t = useTranslation().trans;
  const [open, setOpen] = useState<boolean>(false);
  const [search, setSearch] = useState<string>('');
  const dedbouncedSearch = useDebounced(search, 500);
  const currency = useNumber().currency;

  const computedItemAmount = (qty: number) => {
    setAmount(qty * (currentItem?.price || 0));
  };

  useEffect(() => {
    const searchItems = () => {
      router.reload({ only: ['items'], data: { search: dedbouncedSearch }, preserveUrl: true });
    };
    if (dedbouncedSearch) {
      searchItems();
    }
  }, [dedbouncedSearch]);

  const handleOnKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.currentTarget.name === 'reference') {
      if (event.key === 'k' && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        setOpen(true);
        return;
      }
    }

    handleKeyDown(event);
  };

  const handleCreateNewItem = () => {
    alert('this feature is not implemented yet. please create the item first and then add it to the invoice');
  };

  return (
    <>
      <table className="w-full table-auto">
        <thead>
          <LinesForm
            kind={kind}
            currentItem={currentItem}
            amount={amount}
            currency={currency}
            handleOnKeyDown={handleOnKeyDown}
            computedItemAmount={computedItemAmount}
            referenceInputRef={referenceInputRef}
            qtyInputRef={qtyInputRef}
          />
          <LinesColumnHeaders />
        </thead>
        <tbody>
          {lines &&
            lines
              .filter((l) => l.action !== 'deleted')
              .map((line, index) => <Line key={index} line={line} index={index} currency={currency} handleRemoveLine={handleRemoveLine} />)}
        </tbody>
        {lineError && (
          <tfoot>
            <tr>
              <td colSpan={7}>
                <div className="py-3">
                  <AlertDestructive description={lineError} destroyable={false} />
                </div>
              </td>
            </tr>
          </tfoot>
        )}
      </table>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder={t(`${kind}s.line.form.search`)} value={typeof search === 'string' ? search : ''} onValueChange={setSearch} />
        <Command>
          <CommandList className="min-h-40">
            <CommandGroup className="max-h-60 min-h-40 overflow-y-auto">
              {items &&
                items.map((item) => (
                  <CommandItem
                    asChild
                    value={`${item.id}-${item.variant_id}`}
                    key={`${item.id}-${item.variant_id}`}
                    onSelect={() => {
                      handleOnSelected(item);
                      setOpen(false);
                    }}
                  >
                    <div className="flex w-full items-center justify-between">
                      <div className="flex w-full flex-col items-start justify-start gap-y-0">
                        <div>{item.name}</div>
                        <div className="text-muted-foreground text-xs">{item.description}</div>
                      </div>
                      <div className="text-xl font-medium">{currency(item.price)}</div>
                    </div>
                  </CommandItem>
                ))}
            </CommandGroup>
          </CommandList>
        </Command>
        <div className="flex w-full items-center justify-center rounded-b-lg border bg-gray-100/25 py-2">
          <button className="flex cursor-pointer items-center justify-center gap-x-2 text-indigo-400" onClick={handleCreateNewItem}>
            <LayoutListIcon className="size-4" /> {t(`${kind}s.line.form.addNew`)}
          </button>
        </div>
      </CommandDialog>
    </>
  );
};
