import { Link } from '@inertiajs/react';

export function EmptyChartState() {
  return (
    <div className="flex h-64 flex-col items-center justify-center text-gray-500">
      {/* Illustration */}
      <svg className="mb-4 h-16 w-16 text-gray-400" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 17v-6h6v6h5v2H4v-2h5zM9 7h6v2H9V7z" />
      </svg>

      {/* Message */}
      <p className="text-lg font-semibold">No chart data available</p>
      <p className="mb-4 text-sm text-gray-400">Try selecting another period or adding invoices.</p>

      {/* Call to action */}
      <Link className="bg-primary hover:bg-primary-600 rounded-md px-4 py-2 text-white shadow transition" href={'/invoices/create'}>
        Create Invoice
      </Link>
    </div>
  );
}
