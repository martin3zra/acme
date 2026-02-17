import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps } from '@/types';
import { Transition } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';

interface AccountForm {
  id: string;
  name: string;
  email: string;
}

export default function AccountForm() {
  const t = useTranslation().trans;
  const { auth } = usePage<PageProps>().props;
  const { headers } = useHeader();

  const { data, setData, put, errors, processing, recentlySuccessful } = useForm<Required<AccountForm>>({
    id: auth.user.uuid,
    name: auth.user.name,
    email: auth.user.email,
  });
  const submit = () => {
    put(`/settings/${auth.account.uuid}/profile`, { ...headers });
  };
  return (
    <div>
      {auth.user.pending_email !== null && (
        <div className="col-span-12 space-y-2">
          <Alert variant="destructive" className="border-red-400 bg-red-100/50">
            <AlertDescription
              className="inline"
              dangerouslySetInnerHTML={{ __html: t('profile.alert.emailChangePending', { email: auth.user.pending_email }) }}
            />
          </Alert>
        </div>
      )}
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('global.profile')}</FormSection.Title>
        <FormSection.Description>{t('profile.description')}</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 space-y-2">
            <Label htmlFor="name" className="text-end">
              {t('global.name')}
            </Label>
            <Input
              type="text"
              name="name"
              className="h-12 md:text-xl"
              value={data.name}
              onChange={(e) => setData('name', e.target.value)}
              autoFocus
            />
            <InputError message={errors.name} />
          </div>
          <div className="col-span-6 space-y-2">
            <Label htmlFor="email" className="text-end">
              {t('global.email')}
            </Label>
            <Input type="email" name="email" className="h-12 md:text-xl" value={data.email} onChange={(e) => setData('email', e.target.value)} />
            <InputError message={errors.email} />
          </div>
        </FormSection.Form>
        <FormSection.Actions>
          <Transition
            show={recentlySuccessful}
            enter="transition ease-in-out"
            enterFrom="opacity-0"
            leave="transition ease-in-out"
            leaveTo="opacity-0"
          >
            <p className="text-sm text-gray-600">Saved.</p>
          </Transition>
          <Button type="submit" disabled={processing} className="h-12 md:text-xl">
            {processing && <Spinner />}
            {t('profile.actions.updateProfile')}
          </Button>
        </FormSection.Actions>
      </FormSection>
    </div>
  );
}
