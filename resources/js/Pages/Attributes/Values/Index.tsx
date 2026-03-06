import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { router } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { useState } from 'react';
import { Attribute, AttributeValue } from '../types';
import { List } from './List/Index';
import CreateForm from './Shared/CreateForm';

export default function Index({
  auth,
  attribute,
  values,
}: PageProps<{ attribute: Attribute; values: AttributeValue[] }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [selectedValue, setSelectedValue] = useState<AttributeValue | undefined>(undefined);
  const hasValues = values.length > 0;

  const breadcrumbs = [
    { title: t('global.attributes'), href: '/attributes' },
    { title: attribute.display_name, href: `/attributes/${attribute.uuid}` },
    { title: t('global.actions.title'), href: '' },
  ];

  const handleCreate = () => {
    setSelectedValue(undefined);
    setOpen(true);
  };

  const handleEdit = (value: AttributeValue) => {
    setSelectedValue(value);
    setOpen(true);
  };

  const onOpenChange = (value: boolean) => {
    setOpen(value);
    if (!value) setSelectedValue(undefined);
  };

  const handleDelete = (id: number) => {
    if (confirm(t('global.actions.delete') + '?')) {
      router.delete(`/attribute-values/${id}/${attribute.uuid}`);
    }
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasValues && (
          <HeadingSmall
            title={`${attribute.display_name} ${t('global.attributeValues')}`}
            description={t('attributes.values.description')}
            rightPanel={
              <Button onClick={handleCreate}>
                <Plus /> {t('attributes.values.newValue.title')}
              </Button>
            }
          />
        )}

        {!hasValues ? (
          <div className="absolute top-1/2 left-1/2 flex h-61 min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-3xl bg-white p-10 shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
            <h4 className="text-2xl">{t('global.attributeValues')}</h4>
            <p className="text-sm text-gray-400">{t('attributes.values.emptyState.description')}</p>

            <div className="flex space-x-3">
              <Button onClick={handleCreate}>
                <Plus /> {t('attributes.values.newValue.title')}
              </Button>
            </div>
          </div>
        ) : (
          <List data={values} onEdit={handleEdit} onDelete={handleDelete} t={t} />
        )}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-7xl">
            <SheetHeader>
              <SheetTitle>
                {selectedValue ? t('global.actions.edit') : t('global.actions.create')} {t('global.attributeValue')}
              </SheetTitle>
              <SheetDescription>
                {selectedValue
                  ? `${t('global.update')} ${attribute.display_name} value`
                  : `${t('global.create')} new ${attribute.display_name} value`}
              </SheetDescription>
            </SheetHeader>

            <div className="no-scrollbar grid gap-4 overflow-y-scroll px-4">
              <CreateForm attributeId={attribute.id} value={selectedValue} onFinish={() => onOpenChange(false)} />
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </AppLayout>
  );
}
