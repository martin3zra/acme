import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import AuthLayout from '@/layouts/auth-layout';
import { useForm, usePage } from '@inertiajs/react';
import { FormEventHandler } from 'react';

interface LoginForm {
  email: string;
  password: string;
  remember: boolean;
}

export default function Login() {
  const { data, setData, post, processing, errors, reset } = useForm<Required<LoginForm>>({
    email: '',
    password: '',
    remember: false,
  });

  const props = usePage().props;
  const t = useTranslation().trans;

  const submit: FormEventHandler = (e) => {
    e.preventDefault();
    post('login', {
      headers: {
        'X-CSRF-Token': props.csrf_token as string,
      },
      onFinish: () => reset('password'),
    });
  };

  return (
    <AuthLayout title={t('auth.login.title')} description={t('auth.login.description')}>
      <form className="flex flex-col gap-6" onSubmit={submit}>
        <div className="grid gap-6">
          <div className="grid gap-2">
            <Label htmlFor="email">{t('global.email')}</Label>
            <Input
              id="email"
              type="email"
              autoFocus
              tabIndex={1}
              autoComplete="email"
              value={data.email}
              onChange={(e) => setData('email', e.target.value)}
            />
            <InputError message={errors.email} />
          </div>

          <div className="grid gap-2">
            <div className="flex items-center">
              <Label htmlFor="password">{t('global.password')}</Label>
            </div>
            <Input
              id="password"
              type="password"
              tabIndex={2}
              autoComplete="current-password"
              value={data.password}
              onChange={(e) => setData('password', e.target.value)}
            />
            <InputError message={errors.password} />
          </div>

          <Button type="submit" className="w-full" tabIndex={3} disabled={processing}>
            {processing ? <div>{t('global.processing')}</div> : <span>{t('auth.login.action.login')}</span>}
          </Button>
        </div>
      </form>
    </AuthLayout>
  );
}
