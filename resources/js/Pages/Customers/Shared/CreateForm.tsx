import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import { Customer, PageProps, Verb } from '@/types';
import { useForm, usePage } from '@inertiajs/react';
import { FormEventHandler, useState } from 'react';

export type CreateFormParams = {
  customer: Customer | undefined;
  action: Verb;
};

type CreateFormProps = {
  onFinish: () => void;
  params: CreateFormParams;
};

type CustomerForm = {
  id: number | undefined;
  name: string;
  contact: string;
  email: string;
  phone: string;
};

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { data, setData, post, put, errors, reset, processing } = useForm<Required<CustomerForm>>({
    id: params.customer?.id,
    name: params.customer?.name || '',
    contact: params.customer?.contact_name || '',
    email: params.customer?.email || '',
    phone: params.customer?.phone || '',
  });

  const viewMode = params.action === 'view';
  const isDisabled = params.customer?.status === 'disabled';

  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => {
      reset();
      onFinish();
    },
  };

  const verbName = useVerb().action(params.action);

  const submit: FormEventHandler = (e) => {
    e.preventDefault();

    if (params.action === 'create') post('/customers', options);
    if (params.action === 'edit') put(`/customers/${params.customer!.id}`, options);
  };

  const dialogProps = {
    enabled: {
      title: `Are you sure you want to disable ${params.customer?.name}?`,
      description:
        'Once this customer is disabled, all of its resources and data will also be lock. Please enter your password ' +
        'to confirm you would like to disabled this customer.',
      action: 'Disabled',
      variant: 'destructive',
    },
    disabled: {
      title: `Are you sure you want to enable ${params.customer?.name}?`,
      description:
        'Once this customer is enable, all of its resources and data will also be unlock. Please enter your password ' +
        'to confirm you would like to enabled this customer.',
      action: 'Enabled',
      variant: 'primary',
    },
  }[params.customer?.status || 'enabled'];

  return (
    <div>
      {propsErrors.status && <div className="mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}
      <form onSubmit={submit} className="space-y-6">
        <div className="grid grid-cols-2 gap-2">
          <div className="grid gap-2">
            <Label htmlFor="name">{t('global.name')}</Label>
            <Input
              id="name"
              className="mt-1 block w-full"
              value={data.name}
              onChange={(e) => setData('name', e.target.value)}
              required
              autoComplete="name"
              placeholder={t('global.name')}
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.name} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="contact">{t('global.contact')}</Label>
            <Input
              id="contact"
              className="mt-1 block w-full"
              value={data.contact}
              onChange={(e) => setData('contact', e.target.value)}
              placeholder="Jane Doe"
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.contact} />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2">
          <div className="grid gap-2">
            <Label htmlFor="email">{t('global.email')}</Label>
            <Input
              id="email"
              type="email"
              className="mt-1 block w-full"
              value={data.email}
              onChange={(e) => setData('email', e.target.value)}
              required
              placeholder="jane.doe@example.com"
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.email} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="phone">{t('global.phone')}</Label>
            <Input
              id="phone"
              className="mt-1 block w-full"
              value={data.phone}
              onChange={(e) => setData('phone', e.target.value)}
              placeholder="809-983-3897"
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.phone} />
          </div>
        </div>
        {!viewMode && (
          <div className="flex items-center gap-4">
            <Button disabled={processing}>
              {t(`global.actions.${verbName}`)} {t('global.customer')}
            </Button>
          </div>
        )}
      </form>

      {viewMode && (
        <div className="space-y-6 pt-12">
          <HeadingSmall
            title={`${dialogProps?.action} customer`}
            description={`${isDisabled ? 'Unlock' : 'Lock'} customer and all of its resources`}
          />
          <div className={`space-y-4 rounded-lg border ${isDisabled ? 'border-primary-100 bg-primary-50' : 'border-red-100 bg-red-50'} p-4`}>
            <div className={`relative space-y-0.5 ${isDisabled ? 'text-primary' : 'text-red-600'}`}>
              <p className="font-medium">Warning</p>
              <p className="text-sm">Please proceed with caution, this is not permanent.</p>
            </div>
            <Button variant={isDisabled ? 'default' : 'destructive'} onClick={() => setDialogOpen(true)}>
              {`${dialogProps?.action} customer`}
            </Button>

            <ConfirmsPassword
              title={dialogProps?.title || ''}
              description={dialogProps?.description || ''}
              action={`${dialogProps?.action} it`}
              verb={'update'}
              path={`/customers/${params.customer?.id}/change-status`}
              open={dialogOpen}
              onOpenChange={setDialogOpen}
            />
          </div>
        </div>
      )}
    </div>
  );
}
