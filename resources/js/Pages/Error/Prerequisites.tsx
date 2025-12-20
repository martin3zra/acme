import AppLayout from "@/layouts/app-layout";
import { PageProps } from "@/types";
import { Head, usePage } from "@inertiajs/react";

type MissingItem = {
  key: string;
  message: string;
};

type Props = {
  resource: string;
  missing: MissingItem[];
};

export default function Prerequisites({ resource, missing }: Props) {
  const { auth } = usePage<PageProps>().props;
  return (
    <>
      <Head title={`Cannot continue`} />

      <div className="max-w-2xl mx-auto mt-20 p-6 bg-white rounded-xl shadow">
        <h1 className="text-xl font-semibold text-red-600">
          Action blocked
        </h1>

        <p className="mt-2 text-gray-600">
          Before you can continue with <strong>{resource}</strong>, the following
          items must be configured:
        </p>

        <ul className="mt-4 space-y-2">
          {missing.map(item => (
            <li
              key={item.key}
              className="flex items-start gap-2 p-3 bg-red-50 border border-red-200 rounded"
            >
              <span className="text-red-600">•</span>
              <span>{item.message}</span>
            </li>
          ))}
        </ul>

        <div className="mt-6 flex gap-3">
          <a
            href={`/settings/${auth.account.uuid}/profile`}
            className="px-4 py-2 bg-primary text-primary-foreground rounded"
          >
            Go to settings
          </a>

          <a
            href="/home"
            className="px-4 py-2 border rounded"
          >
            Back to home
          </a>
        </div>
      </div>
    </>
  );
}
