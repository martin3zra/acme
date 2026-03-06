import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps, Verb } from '@/types';
import { Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Attribute, AttributeValue } from '../types';
import { List } from './List/Index';
import CreateForm, { CreateFormParams } from './Shared/CreateForm';

export default function Index({ auth, attribute, values }: PageProps<{ attribute: Attribute; values: AttributeValue[] }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [selectedValue, setSelectedValue] = useState<CreateFormParams>({
    attributeId: attribute.uuid,
    value: undefined,
    action: 'create',
  });
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const hasValues = values.length > 0;

  const breadcrumbs = [
    { title: t('global.attributes'), href: '/attributes' },
    { title: attribute.display_name, href: '' },
    { title: t('attributes.values.title'), href: '' },
  ];

  const handleCreate = () => {
    setSelectedValue({ attributeId: attribute.uuid, value: undefined, action: 'create' });
    setOpen(true);
  };

  const onSelectValue = (value: AttributeValue, action: Verb) => {
    setSelectedValue({ attributeId: attribute.uuid, value, action });
  };

  const onOpenChange = (value: boolean) => {
    setOpen(value);
    if (!value) setSelectedValue({ attributeId: attribute.uuid, value: undefined, action: 'create' });
  };

  useEffect(() => {
    if (selectedValue && selectedValue.value !== undefined) {
      if (selectedValue.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedValue]);

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasValues && (
          <HeadingSmall
            title={`${t('attributes.values.title')} · ${attribute.display_name}`}
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
            <h4 className="text-2xl">{t('attributes.values.emptyState.title')}</h4>
            <p className="text-sm text-gray-400">{t('attributes.values.emptyState.description')}</p>

            <div className="flex space-x-3">
              <Button onClick={handleCreate}>
                <Plus /> {t('attributes.values.newValue.title')}
              </Button>
            </div>
          </div>
        ) : (
          <List data={values} onSelectValue={onSelectValue} t={t} />
        )}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-7xl">
            <SheetHeader>
              <SheetTitle>
                {selectedValue.action === 'edit' ? t('attributes.values.editValue.title') : t('attributes.values.newValue.title')}
              </SheetTitle>
              <SheetDescription>
                {selectedValue.action === 'edit' ? t('attributes.values.editValue.description') : t('attributes.values.newValue.description')}
              </SheetDescription>
            </SheetHeader>

            <div className="no-scrollbar grid gap-4 overflow-y-scroll px-4">
              <CreateForm params={selectedValue} onFinish={() => modalHandler(false)} />
            </div>
          </SheetContent>
        </Sheet>

        {selectedValue.value && (
          <ConfirmsPassword
            title={t('attributes.values.confirmsPassword.title', { value: selectedValue.value.display_name })}
            description={t('attributes.values.confirmsPassword.description')}
            action={t('attributes.values.confirmsPassword.confirm')}
            verb={'destroy'}
            path={`/attribute-values/${selectedValue.value.uuid}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AppLayout>
  );
}
