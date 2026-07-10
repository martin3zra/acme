import ActionSection from '@/components/action-section';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { useHeader } from '@/composables/use-headers';
import { Company, PageProps, Role, User, UserVerb } from '@/types';
import { Field, Radio, RadioGroup, Transition } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';
import { CheckCircleIcon } from 'lucide-react';
import { useState } from 'react';

type CompanyRole = {
  company: string;
  role: string;
};

export interface UserForm {
  name: string;
  email: string;
  companies: CompanyRole[];
}

export type UserFormParams = {
  user: User | undefined;
  action: UserVerb;
};

export type UserFormProps = {
  onFinish: () => void;
  params: UserFormParams;
  companies: Company[];
  roles: Role[];
};

type UserCompany = {
  company: Company;
  role: Role;
};

export default function UserForm({ params, companies, roles }: UserFormProps) {
  const { auth } = usePage<PageProps>().props;
  const [userCompanies, setUserCompanies] = useState<UserCompany[]>((): UserCompany[] => {
    if (params.user === undefined) return [];
    return params.user.linkedCompanies
      .map((linked) => {
        const company = companies.find((c: Company) => c.uuid === linked.uuid);
        const role = roles.find((c: Role) => c.id === linked.role);
        return { company, role } as UserCompany;
      })
      .filter((l): l is UserCompany => l !== null); // ✨ TS type-guard
  });
  const { headers } = useHeader();
  const { data, setData, transform, post, put, errors, processing, recentlySuccessful } = useForm<Required<UserForm>>({
    name: params.user?.name || 'Nathalia',
    email: params.user?.email || 'thalia@example.com',
    companies: [],
  });

  // Add or update business-role selection
  const updateUserCompanies = (company: Company, role: Role) => {
    setUserCompanies((prev) => {
      const exists = prev.find((b: UserCompany) => b.company.id === company.id);
      return exists ? prev.map((b) => (b.company.id === company.id ? { company, role } : b)) : [...prev, { company, role }];
    });
  };

  const options = { ...headers }; //, onSuccess: () => onFinish() };
  const submit = () => {
    transform((data) => ({
      ...data,
      companies: userCompanies.map(({ company, role }) => ({ company: company.uuid, role: role.id })),
    }));

    if (params.action === 'create') {
      post(`/settings/${auth.account.uuid}/users`, { ...options, preserveState: 'errors' });
      return;
    }

    put(`/settings/${auth.account.uuid}/users/${params.user?.uuid}`, { ...options, preserveState: 'errors' });
  };

  return (
    <div className="flex flex-col space-y-6">
      <FormSection onSubmit={submit}>
        <FormSection.Title>Account</FormSection.Title>
        <FormSection.Description>Manage your account settings and set e-mail preferences.</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 space-y-2">
            {params.user !== undefined && params.user?.email_verified_at === null && (
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

          <Separator className="col-span-6" />
          {companies.map((company) => (
            <div className="col-span-6" key={`company-${company.uuid}`}>
              <h3 className="mb-5 text-lg font-medium text-gray-900 dark:text-white">{company.name}</h3>
              <RadioGroup
                by="label"
                aria-label="Companies Roles"
                className="grid grid-cols-3 gap-6"
                onChange={(role: Role) => updateUserCompanies(company, role)}
                value={userCompanies.find((b: UserCompany) => b.company.id === company.id)?.role}
              >
                {roles.map((role) => (
                  <Field key={role.id}>
                    <Radio
                      value={role}
                      className="group data-checked:bg-primary data-checked:text-primary-foreground bg-primary/5 data-focus:outline-primary relative flex cursor-pointer grid-cols-1 rounded-lg px-5 py-4 shadow-md transition focus:not-data-focus:outline-none data-focus:outline"
                    >
                      <div className="flex w-full flex-col items-start justify-start">
                        <div className="text-lg font-semibold">{role.label}</div>
                        <div className="text-xs">{role.description}</div>
                      </div>
                      <CheckCircleIcon className="size-6 opacity-0 transition group-data-checked:opacity-100" />
                    </Radio>
                  </Field>
                ))}
              </RadioGroup>
            </div>
          ))}
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
          <Button type="submit" disabled={processing} className="h-12 uppercase md:text-xl">
            save
          </Button>
        </FormSection.Actions>
      </FormSection>

      {params.user !== undefined && (
        <>
          <Separator className="space-y-6" />
          <ActionSection>
            <ActionSection.Title>Danger Zone: Disable User Account</ActionSection.Title>
            <ActionSection.Description>
              Disabling this user account will immediately revoke their access to the system. The user will be logged out, and they will no longer be
              able to perform any actions or access any resources. This action does not delete the user's data or history but will effectively
              deactivate the account until re-enabled by an administrator.
            </ActionSection.Description>
            <ActionSection.Content>
              <p>
                <span className="font-bold">Security Notice</span> Disabling a user account is a sensitive action. Only do this if the user should no
                longer access the system. For temporary suspensions, consider using role restrictions instead. Need to undo this later? You can
                re-enable a disabled user at any time from the Admin {'>'} Users section. No data will be lost when disabling.
              </p>

              <Button className="mt-6 uppercase" variant={'destructive'}>
                Confirm and Disable
              </Button>
            </ActionSection.Content>
          </ActionSection>
        </>
      )}
    </div>
  );
}
