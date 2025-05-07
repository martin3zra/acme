import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useDocument } from '@/composables/use-document';
import { useHeader } from '@/composables/use-headers';
import { useForm } from '@inertiajs/react';
import { FC, FormEventHandler, useRef } from 'react';
import { AlertDestructive } from './alert-destructive';

export type Props = {
  title: string;
  description: string;
  action: string;
  verb: 'destroy' | 'update';
  path: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export const ConfirmsPassword: FC<Props> = ({ title, description, action, verb, path, open, onOpenChange }) => {
  const { removeElementParent } = useDocument();
  const passwordInput = useRef<HTMLInputElement>(null);
  const {
    data,
    setData,
    delete: destroy,
    put,
    processing,
    reset,
    errors,
    clearErrors,
  } = useForm<Required<{ current_password: string; status: string }>>({ current_password: '', status: '' });
  const { headers } = useHeader();

  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => handleOnOpenChange(),
    onError: () => passwordInput.current?.focus(),
    onFinish: () => reset(),
  };

  const onSubmit: FormEventHandler = (e) => {
    e.preventDefault();

    if (verb == 'destroy') destroy(path, { ...options, preserveState: 'errors' });
    if (verb == 'update') put(path, { ...options, preserveState: 'errors' });
  };

  const handleOnOpenChange = (open: boolean = false) => {
    clearErrors();
    reset();
    onOpenChange(open);
  };

  return (
    <div className="space-y-6">
      <Dialog open={open} onOpenChange={handleOnOpenChange}>
        <DialogContent className="sm:max-w-lg" onInteractOutside={(e) => removeElementParent(e, 'iframe')}>
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            <DialogDescription>{description}</DialogDescription>
          </DialogHeader>
          <form className="space-y-6" onSubmit={onSubmit}>
            {errors.status && <AlertDestructive description={errors.status} onDestroy={() => delete errors.status} />}
            <div className="grid gap-2">
              <Label htmlFor="password" className="sr-only">
                Password
              </Label>

              <Input
                id="current_password"
                type="password"
                name="current_password"
                ref={passwordInput}
                value={data.current_password}
                onChange={(e) => setData('current_password', e.target.value)}
                placeholder="Password"
              />

              <InputError message={errors.current_password} />
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="secondary">
                  Close
                </Button>
              </DialogClose>
              <Button variant="destructive" disabled={processing} asChild>
                <button type="submit">{action}</button>
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
};
