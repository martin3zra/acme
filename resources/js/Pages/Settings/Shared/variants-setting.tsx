import FormSection from '@/components/form-section';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps } from '@/types';
import { useForm, usePage } from '@inertiajs/react';

type Props = {
  uuid: string;
  enabled: boolean;
};

// Toggles the company-level "handles variants" feature flag. When on, the item
// editor exposes the attribute/variant matrix.
export function VariantsSetting({ uuid, enabled }: Props) {
  const t = useTranslation().trans;
  const { auth } = usePage<PageProps>().props;
  const { headers } = useHeader();
  const { data, setData, put, processing } = useForm({ enabled });

  const handleSubmit = () => {
    put(`/settings/${auth.account.uuid}/companies/${uuid}/handles-variants`, { ...headers, preserveState: 'errors' });
  };

  return (
    <div className="flex w-full flex-col space-y-6 py-6 **:data-form:md:grid-cols-1">
      <FormSection onSubmit={handleSubmit}>
        <FormSection.Title>{t('profile.companies.viewCompany.variants.title')}</FormSection.Title>
        <FormSection.Description>{t('profile.companies.viewCompany.variants.description')}</FormSection.Description>
        <FormSection.Form>
          <label className="flex items-center gap-3">
            <Checkbox checked={data.enabled} onCheckedChange={(checked) => setData('enabled', checked === true)} />
            <Label className="cursor-pointer">{t('profile.companies.viewCompany.variants.enable')}</Label>
          </label>
        </FormSection.Form>
        <FormSection.Actions>
          <Button disabled={processing} className="uppercase">
            {t('global.actions.update')}
          </Button>
        </FormSection.Actions>
      </FormSection>
    </div>
  );
}
