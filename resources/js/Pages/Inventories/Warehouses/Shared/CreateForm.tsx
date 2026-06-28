import ActionSection from '@/components/action-section';
import { ConfirmsPassword } from '@/components/confirms-password';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps, Verb, Warehouse } from '@/types';
import { useForm, usePage } from '@inertiajs/react';
import { useState } from 'react';

export type CreateFormParams = {
  warehouse: Warehouse | undefined;
  action: Verb;
};

type CreateFormProps = {
  onFinish: () => void;
  params: CreateFormParams;
};

type WarehouseForm = {
  id: number | undefined;
  name: string;
  location: string;
};

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { headers } = useHeader();

  const { data, setData, post, put, errors, reset, processing } = useForm<Required<WarehouseForm>>({
    id: params.warehouse?.id,
    name: params.warehouse?.name || '',
    location: params.warehouse?.location || '',
  });

  const viewMode = params.action === 'view';
  const isDisabled = params.warehouse?.status === 'disabled';

  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => {
      reset();
      onFinish();
    },
  };

  const verbName = useVerb().action(params.action);

  const submit = () => {
    if (params.action === 'create') post('/inventories/warehouses', options);
    if (params.action === 'edit') put(`/inventories/warehouses/${params.warehouse!.id}`, options);
  };

  return (
    <div className="flex flex-col space-y-6">
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('warehouses.single.title')}</FormSection.Title>
        <FormSection.Description>{t('warehouses.single.description')}</FormSection.Description>
        <FormSection.Form>
          {propsErrors.status && <div className="mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}

          <div className="col-span-6 flex flex-col gap-y-6">
            <div className="flex flex-col gap-2">
              <Label htmlFor="name">{t('global.name')}</Label>
              <Input
                id="name"
                className="mt-1 block w-full"
                value={data.name}
                onChange={(e) => setData('name', e.target.value)}
                required
                autoComplete="name"
                placeholder={t('warehouses.namePlaceholder')}
                readOnly={viewMode}
              />
              <InputError className="mt-2" message={errors.name} />
            </div>

            <div className="flex flex-col gap-2">
              <div className="grid gap-2">
                <Label htmlFor="location">{t('global.location')}</Label>
                <textarea
                  id="location"
                  className="mt-1 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500"
                  value={data.location}
                  onChange={(e) => setData('location', e.target.value)}
                  placeholder={t('warehouses.locationPlaceholder')}
                  rows={3}
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.location} />
              </div>
            </div>
          </div>
        </FormSection.Form>

        {!viewMode && (
          <FormSection.Actions>
            <div className="flex items-center gap-4">
              <Button disabled={processing} className="uppercase">
                {t(`global.actions.${verbName}`)} {t('global.warehouse')}
              </Button>
            </div>
          </FormSection.Actions>
        )}
      </FormSection>

      {viewMode && (
        <ActionSection>
          <ActionSection.Title>{t(`warehouses.statuses.${params.warehouse?.status || 'enabled'}.section.title`)}</ActionSection.Title>
          <ActionSection.Description>{t(`warehouses.statuses.${params.warehouse?.status || 'enabled'}.section.description`)}</ActionSection.Description>
          <ActionSection.Content>
            <div className={`space-y-2 rounded-lg border ${isDisabled ? 'border-primary-100 bg-primary-50' : 'border-red-100 bg-red-50'} p-4`}>
              <div className={`relative space-y-0.5 ${isDisabled ? 'text-primary' : 'text-red-600'}`}>
                <p className="font-medium">{t('global.warning.title')}</p>
                <p className="text-sm">{t('global.warning.description')}</p>
              </div>
              <Button variant={isDisabled ? 'default' : 'destructive'} onClick={() => setDialogOpen(true)}>
                {t(`warehouses.statuses.${params.warehouse?.status || 'enabled'}.section.title`)}
              </Button>

              <ConfirmsPassword
                title={t(`warehouses.statuses.${params.warehouse?.status || 'enabled'}.confirmsPassword.title`, { warehouse: params.warehouse?.name || '' })}
                description={t(`warehouses.statuses.${params.warehouse?.status || 'enabled'}.confirmsPassword.description`)}
                action={t(`warehouses.statuses.${params.warehouse?.status || 'enabled'}.confirmsPassword.confirm`)}
                verb={'update'}
                path={`/inventories/warehouses/${params.warehouse?.id}/change-status`}
                open={dialogOpen}
                onOpenChange={setDialogOpen}
              />
            </div>
          </ActionSection.Content>
        </ActionSection>
      )}
    </div>
  );
}
