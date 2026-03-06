import { Button } from '@/components/ui/button';
import HeadingSmall from '@/components/heading-small';
import { useHeader } from '@/composables/use-headers';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { router } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { useState } from 'react';
import { List } from './List/Index';
import { breadcrumbs } from './constants';
import CreateForm from './Shared/CreateForm';
import { Warehouse } from './types';

export default function Index({ auth, warehouses }: PageProps<{ warehouses: Warehouse[] }>) {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const [open, setOpen] = useState(false);
  const [selectedWarehouse, setSelectedWarehouse] = useState<Warehouse | undefined>(undefined);
  const hasWarehouses = warehouses.length > 0;

  const handleCreate = () => {
    setSelectedWarehouse(undefined);
    setOpen(true);
  };

  const handleEdit = (warehouse: Warehouse) => {
    setSelectedWarehouse(warehouse);
    setOpen(true);
  };

  const onOpenChange = (value: boolean) => {
    setOpen(value);
    if (!value) setSelectedWarehouse(undefined);
  };

  const handleDelete = (id: number) => {
    if (confirm('Are you sure?')) {
      router.delete(`/warehouses/${id}`);
    }
  };

  const handleStatusToggle = (id: number) => {
    router.put(`/warehouses/${id}/change-status`, {}, headers);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasWarehouses && (
          <HeadingSmall
            title={t('warehouses.title')}
            description={t('warehouses.description')}
            rightPanel={
              <Button onClick={handleCreate}>
                <Plus /> {t('warehouses.newWarehouse.title')}
              </Button>
            }
          />
        )}
        {!hasWarehouses ? (
          <div className="absolute top-1/2 left-1/2 flex h-61 min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-3xl bg-white p-10 shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
            <h4 className="text-2xl">{t(`warehouses.emptyState.title`)}</h4>
            <p className="text-sm text-gray-400">{t(`warehouses.emptyState.description`)}</p>

            <div className="flex space-x-3">
              <Button onClick={handleCreate}>
                <Plus /> {t('warehouses.newWarehouse.title')}
              </Button>
            </div>
          </div>
        ) : (
          <List data={warehouses} onEdit={handleEdit} onDelete={handleDelete} onStatusToggle={handleStatusToggle} t={t} />
        )}
        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-7xl">
            <SheetHeader>
              <SheetTitle>
                {selectedWarehouse ? t('global.actions.edit') : t('global.actions.create')} {t('global.warehouse')}
              </SheetTitle>
              <SheetDescription>
                {t('warehouses.newWarehouse.description')}
              </SheetDescription>
            </SheetHeader>

            <div className="no-scrollbar grid gap-4 overflow-y-scroll px-4">
              <CreateForm warehouse={selectedWarehouse} onFinish={() => onOpenChange(false)} />
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </AppLayout>
  );
}
