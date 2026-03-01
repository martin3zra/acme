import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import { currencySignature, Item, TransactionKind } from '@/types';

type Props = {
  kind: TransactionKind;
  amount: number;
  currency: currencySignature;
  currentItem: Item | undefined;
  handleOnKeyDown: (event: React.KeyboardEvent<HTMLInputElement>) => void;
  computedItemAmount: (qty: number) => void;
  referenceInputRef: React.RefObject<HTMLInputElement | null>;
  qtyInputRef: React.RefObject<HTMLInputElement | null>;
};

export default function LinesForm({
  kind,
  referenceInputRef,
  qtyInputRef,
  currentItem,
  amount,
  currency,
  handleOnKeyDown,
  computedItemAmount,
}: Props) {
  const t = useTranslation().trans;
  return (
    <tr>
      <th scope="col" className="w-60 border border-gray-300 pe-1">
        <Input
          name="reference"
          ref={referenceInputRef}
          data-reset={false}
          placeholder={t(`${kind}s.line.form.placeholder`)}
          onKeyDown={handleOnKeyDown}
          className="rounded-none border-none focus-visible:border-none focus-visible:ring-[2px]"
          tabIndex={0}
        />
      </th>
      <th scope="col" className="w-auto border border-gray-300 bg-gray-50 px-1">
        <Label>{currentItem?.description}</Label>
      </th>
      <th scope="col" className="w-36 border border-gray-300 bg-gray-50 px-1">
        <Label>{currentItem?.unit.name}</Label>
      </th>
      <th scope="col" className="w-36 border border-gray-300">
        <Input
          type="number"
          min={1}
          name="qty"
          className="rounded-none border-none text-end focus-visible:border-none focus-visible:ring-[2px]"
          tabIndex={1}
          ref={qtyInputRef}
          onFocus={(e) => computedItemAmount(e.currentTarget.valueAsNumber)}
          onChange={(e) => computedItemAmount(e.target.valueAsNumber)}
          onKeyDown={handleOnKeyDown}
        />
      </th>
      <th scope="col" className="w-36 border border-gray-300 bg-gray-50 px-1 text-end">
        <Label className="block">{currency(currentItem?.price || 0)}</Label>
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-end">
        {amount > 0 ? currency(amount) : ''}
      </th>
      <th scope="col" className="w-6 border border-gray-300 text-end" />
    </tr>
  );
}
