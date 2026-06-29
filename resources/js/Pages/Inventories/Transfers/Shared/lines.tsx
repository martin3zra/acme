import { AlertDestructive } from '@/components/alert-destructive';
import { Button } from '@/components/ui/button';
import { Command, CommandDialog, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { useTranslation } from '@/hooks/use-translation';
import { router } from '@inertiajs/react';
import { LayoutListIcon, XCircleIcon } from 'lucide-react';
import { useEffect, useState } from 'react';

export type TransferItem = {
  id: number;
  name: string;
  description: string;
  reference: string;
  sku: string;
  cost: number;
  unit: { id: number | null; name: string | null };
};

export type TransferLine = TransferItem & { qty: number };

type Props = {
  items?: TransferItem[];
  lines: TransferLine[];
  lineError?: string;
  currentItem?: TransferItem;
  amount: number;
  setAmount: React.Dispatch<React.SetStateAction<number>>;
  handleRemoveLine: (event: React.MouseEvent<HTMLButtonElement>) => void;
  handleKeyDown: (event: React.KeyboardEvent<HTMLInputElement>) => void;
  handleOnSelected: (item: TransferItem) => void;
  referenceInputRef: React.RefObject<HTMLInputElement | null>;
  qtyInputRef: React.RefObject<HTMLInputElement | null>;
};

export const Lines = ({
  items,
  lines,
  lineError,
  currentItem,
  amount,
  setAmount,
  handleRemoveLine,
  handleKeyDown,
  handleOnSelected,
  referenceInputRef,
  qtyInputRef,
}: Props) => {
  const t = useTranslation().trans;
  const currency = useNumber().currency;
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounced(search, 500);

  // Global ⌘K / Ctrl+K opens the product browser regardless of focus.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((v) => !v);
      }
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  useEffect(() => {
    if (debouncedSearch) {
      router.reload({ only: ['items'], data: { search: debouncedSearch }, preserveUrl: true });
    }
  }, [debouncedSearch]);

  const recompute = () => {
    const qty = qtyInputRef.current?.valueAsNumber || 0;
    setAmount(qty * (currentItem?.cost ?? 0));
  };

  return (
    <>
      <table className="w-full table-auto">
        <thead>
          {/* Entry row */}
          <tr>
            <th scope="col" className="w-60 border border-gray-300 pe-1">
              <Input
                name="reference"
                ref={referenceInputRef}
                placeholder={t('transfers.line.referencePlaceholder')}
                onKeyDown={handleKeyDown}
                className="rounded-none border-none focus-visible:border-none focus-visible:ring-[2px]"
                tabIndex={0}
              />
            </th>
            <th scope="col" className="w-auto border border-gray-300 bg-gray-50 px-1 text-start">
              <Label>{currentItem?.description}</Label>
            </th>
            <th scope="col" className="w-36 border border-gray-300 bg-gray-50 px-1 text-start">
              <Label>{currentItem?.unit?.name}</Label>
            </th>
            <th scope="col" className="w-28 border border-gray-300">
              <Input
                type="number"
                min={0}
                step="any"
                name="qty"
                className="rounded-none border-none text-end focus-visible:border-none focus-visible:ring-[2px]"
                tabIndex={1}
                ref={qtyInputRef}
                onChange={recompute}
                onKeyDown={handleKeyDown}
              />
            </th>
            <th scope="col" className="w-28 border border-gray-300 bg-gray-50 px-1 text-end">
              <Label className="block">{currency(currentItem?.cost ?? 0)}</Label>
            </th>
            <th scope="col" className="w-32 border border-gray-300 px-1 text-end">
              {amount > 0 ? currency(amount) : ''}
            </th>
            <th scope="col" className="w-6 border border-gray-300 text-end" />
          </tr>
          {/* Column headers */}
          <tr>
            <th scope="col" className="w-60 border border-gray-300 px-1 text-start">
              {t('transfers.line.code')}
            </th>
            <th scope="col" className="w-auto border border-gray-300 px-1 text-start">
              {t('transfers.line.description')}
            </th>
            <th scope="col" className="w-36 border border-gray-300 px-1 text-start">
              {t('transfers.line.unit')}
            </th>
            <th scope="col" className="w-28 border border-gray-300 px-1 text-end">
              {t('transfers.line.qty')}
            </th>
            <th scope="col" className="w-28 border border-gray-300 px-1 text-end">
              {t('transfers.line.cost')}
            </th>
            <th scope="col" className="w-32 border border-gray-300 px-1 text-end">
              {t('transfers.line.total')}
            </th>
            <th scope="col" className="w-6 border border-gray-300 px-5 text-end" />
          </tr>
        </thead>
        <tbody>
          {lines.map((line, index) => (
            <tr key={index}>
              <td className="border border-gray-300 px-1 text-start">{line.reference || line.sku || line.name}</td>
              <td className="border border-gray-300 px-1 text-start">{line.description}</td>
              <td className="border border-gray-300 px-1 text-start">{line.unit?.name}</td>
              <td className="border border-gray-300 px-1 text-end">{line.qty}</td>
              <td className="border border-gray-300 px-1 text-end">{currency(line.cost)}</td>
              <td className="border border-gray-300 px-1 text-end">{currency(line.qty * line.cost)}</td>
              <td className="border border-gray-300 px-1 text-end">
                <Button
                  variant="link"
                  size="icon"
                  className="text-destructive h-8 w-8 cursor-pointer rounded-full p-0"
                  data-index={index}
                  onClick={handleRemoveLine}
                >
                  <XCircleIcon />
                </Button>
              </td>
            </tr>
          ))}
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
        <CommandInput placeholder={t('transfers.line.search')} value={search} onValueChange={setSearch} />
        <Command>
          <CommandList className="min-h-40">
            <CommandGroup className="max-h-72 min-h-40 overflow-y-auto">
              {items?.map((item) => (
                <CommandItem
                  asChild
                  value={`${item.name} ${item.reference} ${item.sku}`}
                  key={item.id}
                  onSelect={() => {
                    handleOnSelected(item);
                    setOpen(false);
                  }}
                >
                  <div className="flex w-full items-center justify-between">
                    <div className="flex flex-col">
                      <span>{item.name}</span>
                      <span className="text-muted-foreground text-xs">
                        {item.reference || item.sku} {item.description ? `· ${item.description}` : ''}
                      </span>
                    </div>
                    <span className="font-medium">{currency(item.cost)}</span>
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
        <div className="flex w-full items-center justify-center rounded-b-lg border bg-gray-100/25 py-2">
          <span className="text-muted-foreground flex items-center gap-x-2 text-xs">
            <LayoutListIcon className="size-4" /> {t('transfers.line.search')}
          </span>
        </div>
      </CommandDialog>
    </>
  );
};
