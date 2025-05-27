import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { Item, PageProps, Tax, Unit, Verb } from '@/types';
import { Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { breadcrumbs } from './constants';
import { List } from './List/Index';
import CreateForm, { CreateFormParams } from './Shared/CreateForm';

export default function Index({ auth, items, taxes, units }: PageProps<{ items: Item[]; taxes: Tax[]; units: Unit[] }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedItem, setSelectedItem] = useState<CreateFormParams>({
    item: undefined,
    taxes,
    units,
    action: 'create',
  });

  const verbName = useVerb().action(selectedItem.action);
  const hasItems = items.length > 0;

  const onCreateNewItem = () => {
    setSelectedItem({ item: undefined, taxes, units, action: 'create' });
    setOpen(!open);
  };

  const onSelectItem = (item: Item, action: Verb): void => {
    setSelectedItem({ item, taxes, units, action });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) setSelectedItem({ item: undefined, taxes, units, action: 'create' });
  };

  useEffect(() => {
    if (selectedItem && selectedItem.item !== undefined) {
      if (selectedItem.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedItem]);

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
              <Button onClick={onCreateNewItem}>
                <Plus /> {t('items.newItem.title')}
              </Button>
            }
          />
        )}

        {!hasItems && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t('items.emptyState.title')}</h4>
              <p className="text-sm text-gray-400">{t('items.emptyState.description')}</p>
              <Button onClick={onCreateNewItem}>
                <Plus /> {t('items.newItem.title')}
              </Button>
            </div>
          </>
        )}

        {hasItems && <List data={items} onSelectItem={onSelectItem} />}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl">
            <SheetHeader>
              <SheetTitle>
                {t(`global.actions.${verbName}`)} {t('global.item')}
              </SheetTitle>
              <SheetDescription className="text-[12px]">{t(`items.newItem.description`)}</SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 px-4">
              <CreateForm params={selectedItem} onFinish={() => modalHandler(false)} />
            </div>
          </SheetContent>
        </Sheet>

        {selectedItem.item && (
          <ConfirmsPassword
            title={t(`items.confirmsPassword.title`, { item: selectedItem?.item.name })}
            description={t(`items.confirmsPassword.description`, { item: selectedItem?.item?.name })}
            action={t(`items.confirmsPassword.confirm`)}
            verb={'destroy'}
            path={`/items/${selectedItem?.item?.id}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AppLayout>
  );
}
