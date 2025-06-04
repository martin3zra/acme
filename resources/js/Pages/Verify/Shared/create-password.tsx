import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { useForm } from '@inertiajs/react';
import { FormEventHandler } from 'react';

interface CreatePasswordForm {
  password: string;
  password_confirmation: string;
}

export default function CreatePassword() {
  const { headers } = useHeader();
  const t = useTranslation().trans;
  const { data, setData, errors, post, processing } = useForm<Required<CreatePasswordForm>>({
    password: '',
    password_confirmation: '',
  });
  const submit: FormEventHandler = (e) => {
    e.preventDefault();

    post('/password', { ...headers });
  };
  return (
    <form onSubmit={submit} className="space-y-6">
      <div
        className="mb-4 text-base text-gray-600 dark:text-gray-400"
        dangerouslySetInnerHTML={{ __html: t(`verify.create-password-description`).replace(/\n/g, '<br>') }}
      />
      <div className="grid gap-2">
        <Label htmlFor="password">{t(`global.password`)}</Label>
        <Input
          id="password"
          type="password"
          required
          autoFocus
          tabIndex={1}
          autoComplete="password"
          value={data.password}
          onChange={(e) => setData('password', e.target.value)}
        />
        <InputError message={errors.password} />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="password_confirmation">{t(`global.password_confirmation`)}</Label>
        <Input
          id="password_confirmation"
          type="password"
          required
          autoFocus
          tabIndex={1}
          autoComplete="password_confirmation"
          value={data.password_confirmation}
          onChange={(e) => setData('password_confirmation', e.target.value)}
        />
        <InputError message={errors.password_confirmation} />
      </div>
      <div className="mt-4 flex items-center justify-between">
        <Button disabled={processing}>{t(`verify.create-password`)}</Button>
      </div>
    </form>
  );
}
