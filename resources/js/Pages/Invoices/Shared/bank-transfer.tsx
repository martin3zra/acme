import FormSection from '@/components/form-section';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import { BankOperationFormProps } from '@/types';

type BTFormProps = BankOperationFormProps & {};

export const BankTransferFormView = ({ amount, reference, onChange }: BTFormProps) => {
  const t = useTranslation().trans;
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.currentTarget.name === 'amount') {
      onChange(event.currentTarget.valueAsNumber);
      return;
    }

    onChange(event.currentTarget.value);
  };
  return (
    <div>
      <FormSection onSubmit={() => {}}>
        <FormSection.Title>{t('global.paymentMethods.bt.form.title')}</FormSection.Title>
        <FormSection.Description>{t('global.paymentMethods.bt.form.description')}</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 space-y-2 sm:col-span-4">
            <Label htmlFor="amount" className="text-end">
              {t('global.amount')}
            </Label>
            <Input type="number" min={0} name="amount" className="h-12 text-end md:text-xl" onChange={handleChange} autoFocus value={amount} />
          </div>
          <div className="col-span-6 space-y-2 sm:col-span-4">
            <Label htmlFor="reference">{t('global.paymentMethods.bt.form.reference')}</Label>
            <Input type="text" name="reference" className="h-12 text-start md:text-xl" onChange={handleChange} value={reference} />
          </div>
        </FormSection.Form>
      </FormSection>
    </div>
  );
};
