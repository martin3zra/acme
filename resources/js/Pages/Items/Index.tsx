import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { BreadcrumbItem, Item, PageProps, Tax, Unit, Verb } from '@/types';
import { Deferred } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { List } from './List/Index';
import CreateForm, { CreateFormParams } from './Shared/CreateForm';

const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Home',
    href: '/home',
  },
  {
    title: 'Items',
    href: '/items',
  },
];

export default function Index({ auth, items, taxes, units }: PageProps<{ items: Item[]; taxes: Tax[]; units: Unit[] }>) {
  const [open, setOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [setSelectedItem, setSelectedItemsetSelectedItem] = useState<CreateFormParams>({
    item: undefined,
    taxes,
    units,
    action: 'create',
  });

  const verbName = useVerb().action(setSelectedItem.action);
  const hasItems = items.length > 0;

  const onCreateNewItem = () => {
    setSelectedItemsetSelectedItem({ item: undefined, taxes, units, action: 'create' });
    setOpen(!open);
  };

  const onSelectItem = (item: Item, action: Verb): void => {
    setSelectedItemsetSelectedItem({ item, taxes, units, action });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) setSelectedItemsetSelectedItem({ item: undefined, taxes, units, action: 'create' });
  };

  useEffect(() => {
    if (setSelectedItem && setSelectedItem.item !== undefined) {
      if (setSelectedItem.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [setSelectedItem]);

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasItems && (
          <HeadingSmall
            title="Items"
            description="All created items are shown here."
            rightPanel={
              <Button onClick={onCreateNewItem}>
                <Plus /> Add items
              </Button>
            }
          />
        )}

        {!hasItems && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">Create your first item</h4>
              <p className="text-sm text-gray-400">Once you create your item, it will appear here.</p>
              <Button onClick={onCreateNewItem}>+ Create Item</Button>
            </div>
          </>
        )}

        {hasItems && <List data={items} onSelectItem={onSelectItem} />}

        <Deferred data={['units', 'taxes']} fallback={<div>Loading...</div>}>
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl">
              <SheetHeader>
                <SheetTitle>{verbName} Item</SheetTitle>
                <SheetDescription className="text-[12px]">Create a new item</SheetDescription>
              </SheetHeader>
              <div className="grid gap-4 px-4">
                <CreateForm params={setSelectedItem} onFinish={() => modalHandler(false)} />
              </div>
            </SheetContent>
          </Sheet>
        </Deferred>

        {setSelectedItem.item && (
          <ConfirmsPassword
            title={`Are you sure you want to delete ${setSelectedItem?.item?.name}?`}
            description={`Once ${setSelectedItem?.item?.name} is deleted, all of its resources will continue to be available, but no new operation can be performed.`}
            action={`Delete it`}
            verb={'destroy'}
            path={`/items/${setSelectedItem?.item?.id}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AuthenticatedLayout>
  );
}
