import FormSection from '@/components/form-section';
import { MoneyInput } from '@/components/money-input';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import { BankOperationFormProps } from '@/types';

export const CheckFormView = ({ amount, reference, onChange }: BankOperationFormProps) => {
  const t = useTranslation().trans;
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange(event.currentTarget.value);
  };
  return (
    <div>
      <FormSection onSubmit={() => {}}>
        <FormSection.Title>{t('global.paymentMethods.ck.form.title')}</FormSection.Title>
        <FormSection.Description>{t('global.paymentMethods.ck.form.description')}</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 space-y-2 sm:col-span-4">
            <Label htmlFor="ck" className="text-end">
              {t('global.amount')}
            </Label>
            <MoneyInput className="h-12 text-end md:text-xl" onChange={(c) => onChange(c)} autoFocus value={amount || 0} />
          </div>
          <div className="col-span-6 space-y-2 sm:col-span-4">
            <Label htmlFor="ck">{t('global.paymentMethods.ck.form.reference')}</Label>
            <Input type="text" name="reference" className="h-12 text-start md:text-xl" onChange={handleChange} value={reference} />
          </div>
        </FormSection.Form>
      </FormSection>
    </div>
  );
};
