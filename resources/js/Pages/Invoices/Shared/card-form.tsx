import FormSection from '@/components/form-section';
import { MoneyInput } from '@/components/money-input';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { defaultCardBrands } from '@/constants';
import { useTranslation } from '@/hooks/use-translation';
import { CardFormInput, PaymentFormType } from '@/types';

type CardFormProps = PaymentFormType & {
  last4: number;
  brand: string;
  onChange: (value: number | string, key: CardFormInput) => void;
};

export const CardFormView = ({ last4, brand, amount, reference, onChange }: CardFormProps) => {
  const t = useTranslation().trans;
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.currentTarget.name === 'last4') {
      event.currentTarget.value = event.currentTarget.value.replace(/\D/g, '');
      if (event.currentTarget.value.length > event.currentTarget.maxLength) {
        event.currentTarget.value = event.currentTarget.value.slice(0, event.currentTarget.maxLength);
      }
      onChange(event.currentTarget.valueAsNumber, event.currentTarget.name);
      return;
    }

    onChange(event.currentTarget.value, event.currentTarget.name as CardFormInput);
  };

  return (
    <div>
      <FormSection onSubmit={() => {}}>
        <FormSection.Title>{t('global.paymentMethods.card.form.title')}</FormSection.Title>
        <FormSection.Description>{t('global.paymentMethods.card.form.description')}</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 space-y-2 sm:col-span-3">
            <Label htmlFor="last4" className="text-end">
              {t('global.paymentMethods.card.form.last4')}
            </Label>
            <Input
              type="number"
              inputMode="numeric"
              name="last4"
              pattern="[0-9]*"
              maxLength={4}
              className="h-12 text-end md:text-xl"
              onChange={handleChange}
              autoFocus
              value={last4}
            />
          </div>
          <div className="col-span-6 space-y-2 sm:col-span-3">
            <Label htmlFor="brand" className="text-end">
              {t('global.paymentMethods.card.form.brand')}
            </Label>
            <Select name="brand" onValueChange={(value) => onChange(value, 'brand')} value={brand} required>
              <SelectTrigger className="w-full" size={'lg'}>
                <SelectValue placeholder="Select brand" />
              </SelectTrigger>
              <SelectContent className="w-full">
                {defaultCardBrands.map((brand, index) => (
                  <SelectItem key={index.toString()} value={brand.value.toString()}>
                    {brand.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="col-span-6 space-y-2 sm:col-span-3">
            <Label htmlFor="reference">{t('global.paymentMethods.card.form.authorization')}</Label>
            <Input type="text" name="reference" className="h-12 text-start md:text-xl" onChange={handleChange} value={reference} />
          </div>
          <div className="col-span-6 space-y-2 sm:col-span-3">
            <Label htmlFor="amount" className="text-end">
              {t('global.amount')}
            </Label>
            <MoneyInput
              name="amount"
              pattern="[0-9]*"
              className="h-12 text-end md:text-xl"
              onChange={(c) => onChange(c, 'amount')}
              value={amount}
            />
          </div>
        </FormSection.Form>
      </FormSection>
    </div>
  );
};
