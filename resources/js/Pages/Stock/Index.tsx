import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { Plus } from 'lucide-react';
import { useState } from 'react';
import { breadcrumbs } from './constants';
import { List } from './List/Index';
import CreateForm from './Shared/CreateForm';
import { StockLevel, Warehouse } from './types';

export default function Index({ auth, stocks, warehouses }: PageProps<{ stocks: StockLevel[]; warehouses: Warehouse[] }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const hasStocks = stocks.length > 0;

  const onOpenChange = (value: boolean) => {
    setOpen(value);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasStocks && (
          <HeadingSmall
            title={t('@global.stockLevels')}
            rightPanel={
              <Button onClick={() => onOpenChange(true)}>
                <Plus /> {t('@global.adjustStock')}
              </Button>
            }
          />
        )}

        {!hasStocks ? (
          <div className="absolute top-1/2 left-1/2 flex h-61 min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-3xl bg-white p-10 shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
            <h4 className="text-2xl">{t('@global.stockLevels')}</h4>
            <p className="text-sm text-gray-400">{t('@global.noDataAvailable')}</p>

            <div className="flex space-x-3">
              <Button onClick={() => onOpenChange(true)}>
                <Plus /> {t('@global.adjustStock')}
              </Button>
            </div>
          </div>
        ) : (
          <List data={stocks} t={t} />
        )}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-7xl">
            <SheetHeader>
              <SheetTitle>{t('@global.adjustStock')}</SheetTitle>
              <SheetDescription>{t('@global.adjustStockQuantity')} for a warehouse and variant</SheetDescription>
            </SheetHeader>

            <div className="no-scrollbar grid gap-4 overflow-y-scroll px-4">
              <CreateForm warehouses={warehouses} onFinish={() => onOpenChange(false)} />
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </AppLayout>
  );
}
