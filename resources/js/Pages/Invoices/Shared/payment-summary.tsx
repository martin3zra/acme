import { useNumber } from '@/composables/use-number';
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

const PaymentSummary: React.FC<PaymentSummaryProps> = ({ paymentData }) => {
  const { usedMethods, primaryMethod } = categorizePayments(paymentData);
  const { currency } = useNumber();

  return (
    <div className="mx-auto w-full rounded-md bg-white">
      <ul className="space-y-1">
        {usedMethods.map((method, index) => (
          <li key={index} className="rounded-md border border-gray-100 bg-gray-50 p-3 text-gray-700">
            <div className="font-medium text-indigo-600 uppercase">{method.method}</div>
            <div>
              <span className="font-semibold">${method.amount.toFixed(2)}</span>
              {method.reference && <span className="ml-2 text-sm text-gray-500">(Ref: {method.reference})</span>}
            </div>
            {method.additionalInfo && (
              <div className="text-sm text-gray-500">
                {method.additionalInfo.brand && <span>Brand: {method.additionalInfo.brand}, </span>}
                {method.additionalInfo.last4 && <span>Last4: {method.additionalInfo.last4}</span>}
              </div>
            )}
          </li>
        ))}
      </ul>
      {primaryMethod && (
        <div className="mt-2 rounded-md border border-indigo-200 bg-indigo-50 p-4">
          <p className="font-semibold text-indigo-700">
            Primary Method: {primaryMethod.method.toUpperCase()}
            <span className="block">{currency(primaryMethod.amount)}</span>
          </p>
        </div>
      )}
    </div>
  );
};

export default PaymentSummary;
