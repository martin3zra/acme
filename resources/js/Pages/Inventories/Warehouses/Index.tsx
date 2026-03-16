import HeadingSmall from '@/components/heading-small';
import { ConfirmsPassword } from '@/components/confirms-password';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps, Verb, Warehouse } from '@/types';
import { useEffect, useState } from 'react';
import { Plus } from 'lucide-react';
import { breadcrumbs } from './constants';
import { List } from './List/Index';
import CreateForm, { CreateFormParams } from './Shared/CreateForm';

export default function Index({
  auth,
  warehouses,
  openState,
}: PageProps<{ openState: boolean; warehouses: Warehouse[] }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState<boolean>(openState);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [selectedWarehouse, setSelectedWarehouse] = useState<CreateFormParams>({
    warehouse: undefined,
    action: 'create',
  });

  const verbName = useVerb().action(selectedWarehouse.action);
  const hasWarehouses = warehouses.length > 0;

  const onCreateNewWarehouse = () => {
    setSelectedWarehouse({ warehouse: undefined, action: 'create' });
    setOpen(true);
  };

  const onSelectWarehouse = (warehouse: Warehouse, action: Verb): void => {
    setSelectedWarehouse({ warehouse, action });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) setSelectedWarehouse({ warehouse: undefined, action: 'create' });
  };

  useEffect(() => {
    if (selectedWarehouse.warehouse !== undefined) {
      if (selectedWarehouse.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedWarehouse]);

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasWarehouses && (
          <HeadingSmall
            title={t('warehouses.title')}
            description={t('warehouses.description')}
            rightPanel={
              <div className="flex space-x-2">
                <Button onClick={onCreateNewWarehouse}>
                  <Plus /> {t('warehouses.newWarehouse.title')}
                </Button>
              </div>
            }
          />
        )}

        {!hasWarehouses && (
          <div className="absolute top-1/2 left-1/2 flex h-61 min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-3xl bg-white p-10 shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
            <h4 className="text-2xl">{t('warehouses.emptyState.title')}</h4>
            <p className="text-sm text-gray-400">{t('warehouses.emptyState.description')}</p>
            <Button onClick={onCreateNewWarehouse}>
              <Plus /> {t('warehouses.newWarehouse.title')}
            </Button>
          </div>
        )}

        {hasWarehouses && <List data={warehouses} onSelectWarehouse={onSelectWarehouse} />}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-5xl">
            <SheetHeader>
              <SheetTitle>
                {t(`global.actions.${verbName}`)} {t('global.warehouse')}
              </SheetTitle>
              <SheetDescription className="text-[12px]">{t('warehouses.single.description')}</SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 overflow-y-scroll px-4">
              <CreateForm params={selectedWarehouse} onFinish={() => modalHandler(false)} />
            </div>
          </SheetContent>
        </Sheet>

        {selectedWarehouse.warehouse && (
          <ConfirmsPassword
            title={t(`warehouses.confirmsPassword.title`, { warehouse: selectedWarehouse.warehouse.name })}
            description={t(`warehouses.confirmsPassword.description`, { warehouse: selectedWarehouse.warehouse.name })}
            action={t(`warehouses.confirmsPassword.confirm`)}
            verb={'destroy'}
            path={`/inventories/warehouses/${selectedWarehouse.warehouse.id}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AppLayout>
  );
}
