import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { ImportDrawer } from '@/components/import-zone';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { Item, ItemTypeFilter, PageProps, Tax, Unit, Verb } from '@/types';
import { Deferred, router, usePage } from '@inertiajs/react';
import { FileUp, Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { breadcrumbs } from './constants';
import { List } from './List/Index';
import CreateForm, { CreateFormParams, ItemAttributeOption } from './Shared/CreateForm';

export default function Index({
  auth,
  items,
  item,
  selectedAction,
  taxes,
  units,
  attributes,
  currentItemTypeFilter,
  openState,
}: PageProps<{
  openState: boolean;
  items: Item[];
  item?: Item;
  selectedAction?: Verb;
  taxes: Tax[];
  units: Unit[];
  attributes: ItemAttributeOption[];
  currentItemTypeFilter: ItemTypeFilter;
}>) {
  const t = useTranslation().trans;
  const page = usePage<PageProps>();
  const [loadingItem, setLoadingItem] = useState<boolean>(false);
  const [open, setOpen] = useState<boolean>(item !== undefined || openState);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [importSheetOpen, setImportSheetOpen] = useState<boolean>(false);
  const [selectedItem, setSelectedItem] = useState<CreateFormParams>({
    item,
    taxes,
    units,
    attributes,
    action: item !== undefined ? selectedAction || 'view' : 'create',
  });

  const verbName = useVerb().action(selectedItem.action);
  const hasItems = items.length > 0;

  const onCreateNewItem = () => {
    setSelectedItem({ item: undefined, taxes, units, attributes, action: 'create' });
    setOpen(true);
  };

  const findSelectedItem = (uuid: string, action: Verb) => {
    router.visit(page.url, {
      except: ['items'],
      data: { id: uuid, action, itemType: currentItemTypeFilter },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingItem(true),
      onFinish: () => setLoadingItem(false),
    });
  };

  const onSelectItem = (item: Item, action: Verb): void => {
    setSelectedItem({ item, taxes, units, attributes, action });

    if (action === 'trash') {
      setDeleteDialogOpen(true);
      findSelectedItem(item.uuid, action);
      return;
    }

    setOpen(true);
    findSelectedItem(item.uuid, action);
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      setSelectedItem({ item: undefined, taxes, units, attributes, action: 'create' });
      const nextURL = currentItemTypeFilter === 'all' ? window.location.pathname : `${window.location.pathname}?itemType=${currentItemTypeFilter}`;
      router.replace({
        url: nextURL,
        preserveScroll: true,
        preserveState: true,
      });
    }
  };

  useEffect(() => {
    if (item === undefined) return;
    setSelectedItem((current) => ({
      ...current,
      item,
      action: selectedAction || current.action || 'view',
    }));
  }, [item, selectedAction]);

  useEffect(() => {
    if (selectedItem && selectedItem.item !== undefined) {
      if (selectedItem.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedItem]);

  const onItemFilterTypeChange = (value: ItemTypeFilter) => {
    router.visit(page.url, {
      data: { itemType: value },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingItem(true),
      onFinish: () => setLoadingItem(false),
    });
  };

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasItems && (
          <HeadingSmall
            title={t('items.title')}
            description={t('items.description')}
            rightPanel={
              <Deferred data={open ? [] : ['taxes', 'units', 'attributes']} fallback={<div>Loading...</div>}>
                <div className="flex space-x-2">
                  <Button onClick={onCreateNewItem}>
                    <Plus /> {t('items.newItem.title')}
                  </Button>
                  <Button onClick={() => setImportSheetOpen(true)}>
                    <FileUp /> {t('global.actions.import')}
                  </Button>
                </div>
              </Deferred>
            }
          />
        )}

        {!hasItems && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-61 min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-3xl bg-white p-10 shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t(`items.emptyState.${currentItemTypeFilter}.title`)}</h4>
              <p className="text-sm text-gray-400">{t(`items.emptyState.${currentItemTypeFilter}.description`)}</p>
              <Deferred data={open ? [] : ['taxes', 'units', 'attributes']} fallback={<div>Loading...</div>}>
                <div className="flex space-x-3">
                  {currentItemTypeFilter !== 'all' && (
                    <Button variant={'outline'} onClick={() => onItemFilterTypeChange('all')}>
                      {t('items.viewAll')}
                    </Button>
                  )}
                  <Button onClick={onCreateNewItem}>
                    <Plus /> {t('items.newItem.title')}
                  </Button>
                  <Button onClick={() => setImportSheetOpen(true)}>
                    <FileUp /> {t('global.actions.import')}
                  </Button>
                </div>
              </Deferred>
            </div>
          </>
        )}

        {hasItems && (
          <List
            data={items}
            currentItemTypeFilter={currentItemTypeFilter}
            onSelectItem={onSelectItem}
            onItemTypeFilterChanges={onItemFilterTypeChange}
          />
        )}

        {!loadingItem && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-7xl">
              <SheetHeader>
                <SheetTitle>
                  {t(`global.actions.${verbName}`)} {t('global.item')}
                </SheetTitle>
                <SheetDescription className="text-[12px]">{t(`items.newItem.description`)}</SheetDescription>
              </SheetHeader>
              <div className="grid gap-4 overflow-y-scroll px-4">
                <CreateForm
                  key={`${selectedItem.action}-${selectedItem.item?.id || 'new'}`}
                  params={selectedItem}
                  onFinish={() => modalHandler(false)}
                />
              </div>
            </SheetContent>
          </Sheet>
        )}

        {selectedItem.item && (
          <ConfirmsPassword
            title={t(`items.confirmsPassword.title`, { item: selectedItem?.item.name })}
            description={t(`items.confirmsPassword.description`, { item: selectedItem?.item?.name })}
            action={t(`items.confirmsPassword.confirm`)}
            verb={'destroy'}
            path={`/items/${selectedItem?.item?.uuid}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
        <ImportDrawer source="items" openImportDrawer={importSheetOpen} setImportDrawer={setImportSheetOpen} />
      </div>
    </AppLayout>
  );
}
