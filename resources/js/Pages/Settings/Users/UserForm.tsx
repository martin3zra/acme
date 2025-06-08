import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { PageProps, User, UserVerb } from '@/types';
import { Transition } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';

export interface UserForm {
  name: string;
  email: string;
}

export type UserFormParams = {
  user: User | undefined;
  action: UserVerb;
};

export type UserFormProps = {
  onFinish: () => void;
  params: UserFormParams;
};

export default function UserForm({ onFinish, params }: UserFormProps) {
  const { auth } = usePage<PageProps>().props;
  const { headers } = useHeader();
  const { data, setData, post, put, errors, processing, recentlySuccessful } = useForm<Required<UserForm>>({
    name: params.user?.name || '',
    email: params.user?.email || '',
  });

  const options = { ...headers }; //, onSuccess: () => onFinish() };
  const submit = () => {
    if (params.action === 'create') {
      post(`/settings/${auth.account.uuid}/users`, { ...options, preserveState: 'errors' });
      return;
    }

    put(`/settings/${auth.account.uuid}/users/${params.user?.uuid}`, { ...options, preserveState: 'errors' });
  };
  return (
    <div>
      <FormSection onSubmit={submit}>
        <FormSection.Title>Account</FormSection.Title>
        <FormSection.Description>Manage your account settings and set e-mail preferences.</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 space-y-2">
            {params.user !== undefined && params.user?.email_verified_at !== null && (
              <Alert variant="destructive" className="border-red-400 bg-red-100/50">
                <AlertDescription className="inline">
                  <span className="font-bold">{params.user?.name}</span> has not verified his account yet.
                </AlertDescription>
              </Alert>
            )}
          </div>
          <div className="col-span-6 space-y-2">
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
          <div className="col-span-6 space-y-2">
            <Label htmlFor="email" className="text-end">
              Email
            </Label>
            <Input
              type="email"
              name="email"
              className="h-12 md:text-xl"
              value={data.email}
              onChange={(e) => setData('email', e.target.value)}
              disabled={params.action !== 'create'}
            />
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
            {params.action}
          </Button>
        </FormSection.Actions>
      </FormSection>
    </div>
  );
}
