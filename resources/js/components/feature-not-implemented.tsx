import { Construction } from 'lucide-react';

export function FeatureNotImplemented() {
  return (
    <div className="flex items-center justify-center py-20">
      <div className="w-full max-w-md rounded-xl border bg-white p-8 text-center shadow-sm">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-gray-100">
          <Construction className="h-6 w-6 text-gray-600" />
        </div>

        <h3 className="text-lg font-semibold text-gray-900">Feature not available yet</h3>

        <p className="mt-2 text-sm text-gray-500">
          This section is planned but hasn't been implemented yet. It will be available in a future update.
        </p>

        <div className="mt-6">
          <span className="inline-flex rounded-md bg-gray-100 px-3 py-1 text-xs font-medium text-gray-700">Coming soon</span>
        </div>
      </div>
    </div>
  );
}
