import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import AppLayout from '@/layouts/app-layout';
import SettingsLayout from '@/layouts/settings/layout';
import { BreadcrumbItem, PageProps } from '@/types';
import { Transition } from '@headlessui/react';
import { Head, useForm, usePage } from '@inertiajs/react';
const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Account Settings',
    href: '',
  },
];

interface AccountForm {
  name: string;
  email: string;
}

export default function Account() {
  const { auth } = usePage<PageProps>().props;
  const { headers } = useHeader();

  const { data, setData, put, errors, processing, recentlySuccessful } = useForm<Required<AccountForm>>({
    name: auth.user.name,
    email: auth.user.email,
  });
  const submit = () => {
    put(`/settings/${auth.account.uuid}/profile`, { ...headers });
  };
  return (
    <AppLayout breadcrumbs={breadcrumbs} user={auth.user}>
      <Head title="Account Settings" />
      <SettingsLayout>
        <FormSection onSubmit={submit}>
          <FormSection.Title>Account</FormSection.Title>
          <FormSection.Description>Manage your account settings and set e-mail preferences.</FormSection.Description>
          <FormSection.Form>
            <div className="col-span-6 space-y-2 sm:col-span-4">
              <Label htmlFor="name" className="text-end">
                Name
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
            <div className="col-span-6 space-y-2 sm:col-span-4">
              <Label htmlFor="email" className="text-end">
                Email
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
              Update account
            </Button>
          </FormSection.Actions>
        </FormSection>
      </SettingsLayout>
    </AppLayout>
  );
}
