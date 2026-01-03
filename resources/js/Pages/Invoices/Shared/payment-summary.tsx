import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import { PaymentMethod } from '@/types';
import React from 'react';

interface PaymentData {
  amount: number;
  reference?: string;
  last4?: number;
  brand?: string;
}

interface PaymentInput {
  cash: PaymentData;
  ck: PaymentData;
  card: PaymentData;
  bt: PaymentData;
}

interface CategorizedPayment {
  method: PaymentMethod;
  amount: number;
  reference?: string;
  additionalInfo?: Record<string, unknown>;
}

interface CategorizedResult {
  usedMethods: CategorizedPayment[];
  primaryMethod: CategorizedPayment | null;
}

const categorizePayments = (data: PaymentInput): CategorizedResult => {
  const methods: PaymentMethod[] = ['cash', 'ck', 'card', 'bt'];

  const usedMethods: CategorizedPayment[] = methods
    .map((method) => {
      const entry = data[method];
      if (entry.amount > 0) {
        const categorized: CategorizedPayment = {
          method,
          amount: entry.amount,
          reference: entry.reference || undefined,
        };
        if (method === 'card') {
          categorized.additionalInfo = {
            brand: entry.brand,
            last4: entry.last4,
          };
        }
        return categorized;
      }
      return null;
    })
    .filter((item): item is CategorizedPayment => item !== null);

  const primaryMethod = usedMethods.reduce<CategorizedPayment | null>((prev, current) => {
    return !prev || current.amount > prev.amount ? current : prev;
  }, null);

  return {
    usedMethods,
    primaryMethod,
  };
};

interface PaymentSummaryProps {
  paymentData: PaymentInput;
}

type CardAdditionalInfo = {
  brand: string
  last4: string
}

const PaymentSummary: React.FC<PaymentSummaryProps> = ({ paymentData }) => {
  const t = useTranslation().trans;
  const { usedMethods, primaryMethod } = categorizePayments(paymentData);
  const { currency } = useNumber();

  function isCardInfo(info: Record<string, unknown>): info is CardAdditionalInfo {
    return (
      typeof info.brand === 'string' &&
      typeof info.last4 === 'number'
    )
  }
  const additionalInfoRenderers: Record<PaymentMethod, (info: Record<string, unknown>) => React.ReactNode> = {
    card: info =>
      isCardInfo(info) ? (
        <div className="text-sm text-gray-500">
          <span>Brand: {info.brand}</span>
          <span className="ml-2">Last4: {info.last4}</span>
        </div>
      ) : null,

      ck: () => null,
      bt: () => null,
      cash: () => null,

  }

  return (
    <div className="mx-auto w-full rounded-md bg-white">
      <ul className="space-y-1">
        {usedMethods.map((method, index) => (
          <li key={index} className="rounded-md border border-gray-100 bg-gray-50 p-3 text-gray-700">
            <div className="font-medium text-indigo-600 uppercase">{t(`global.paymentMethods.${method.method}.title`).toUpperCase()}</div>
            <div>
              <span className="font-semibold">{currency(method.amount)}</span>
              {method.reference && <span className="ml-2 text-sm text-gray-500">(Ref: {method.reference})</span>}
            </div>
            {method.additionalInfo &&
              additionalInfoRenderers[method.method]?.(
                method.additionalInfo
              )
            }
          </li>
        ))}
      </ul>
      {primaryMethod && (
        <div className="mt-2 rounded-md border border-indigo-200 bg-indigo-50 p-4">
          <p className="font-semibold text-indigo-700">
            {t('global.paymentMethods.primary')}: {t(`global.paymentMethods.${primaryMethod.method}.title`).toUpperCase()}
            <span className="block">{currency(primaryMethod.amount)}</span>
          </p>
        </div>
      )}
    </div>
  );
};

export default PaymentSummary;
