import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { router } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { useState } from 'react';
import { breadcrumbs } from './constants';
import { List } from './List/Index';
import CreateForm from './Shared/CreateForm';
import { Attribute } from './types';

export default function Index({ auth, attributes }: PageProps<{ attributes: Attribute[] }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [selectedAttribute, setSelectedAttribute] = useState<Attribute | undefined>(undefined);
  const hasAttributes = attributes.length > 0;

  const handleCreate = () => {
    setSelectedAttribute(undefined);
    setOpen(true);
  };

  const handleEdit = (attribute: Attribute) => {
    setSelectedAttribute(attribute);
    setOpen(true);
  };

  const onOpenChange = (value: boolean) => {
    setOpen(value);
    if (!value) setSelectedAttribute(undefined);
  };

  const handleDelete = (id: number) => {
    if (confirm('Are you sure?')) {
      router.delete(`/attributes/${id}`);
    }
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasAttributes && (
          <HeadingSmall
            title={t('@global.attributes')}
            rightPanel={
              <Button onClick={handleCreate}>
                <Plus /> {t('@global.create')}
              </Button>
            }
          />
        )}

        {!hasAttributes ? (
          <div className="absolute top-1/2 left-1/2 flex h-61 min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-3xl bg-white p-10 shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
            <h4 className="text-2xl">{t('@global.attributes')}</h4>
            <p className="text-sm text-gray-400">{t('@global.noDataAvailable')}</p>

            <div className="flex space-x-3">
              <Button onClick={handleCreate}>
                <Plus /> {t('@global.create')}
              </Button>
            </div>
          </div>
        ) : (
          <List data={attributes} onEdit={handleEdit} onDelete={handleDelete} t={t} />
        )}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-7xl">
            <SheetHeader>
              <SheetTitle>
                {selectedAttribute ? t('@global.actions.edit') : t('@global.actions.create')} {t('@global.attribute')}
              </SheetTitle>
              <SheetDescription>
                {selectedAttribute ? `${t('@global.update')} attribute settings` : `${t('@global.create')} a new attribute`}
              </SheetDescription>
            </SheetHeader>

            <div className="no-scrollbar grid gap-4 overflow-y-scroll px-4">
              <CreateForm attribute={selectedAttribute} onFinish={() => onOpenChange(false)} />
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </AppLayout>
  );
}
