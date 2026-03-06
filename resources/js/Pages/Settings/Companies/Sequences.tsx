import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps, Replacements, SequenceConfig, Sequences } from '@/types';
import { useForm, usePage } from '@inertiajs/react';

type SeqProps = {
  uuid: string;
  sequences: Sequences;
};

export default function SequenceView({ uuid, sequences }: SeqProps) {
  const { auth } = usePage<PageProps>().props;
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { data, setData, transform, put, processing, errors } = useForm({
    sequences: sequences,
  });

  const updateField = (module: string, type: string | null, field: keyof SequenceConfig, value: string | number) => {
    setData((prev) => {
      const updated: Sequences = structuredClone(prev.sequences);

      if (type) {
        const moduleEntry = (updated[module] ?? {}) as Record<string, SequenceConfig>;
        const config = moduleEntry[type] ?? {};

        updated[module] = {
          ...moduleEntry,
          [type]: {
            ...config,
            [field]: value,
          },
        };
      } else {
        updated[module] = {
          ...(updated[module] as SequenceConfig),
          [field]: value,
        };
      }
      console.log(updated);
      return {
        ...prev,
        sequences: updated,
      };
    });
  };

  const handleSave = async () => {
    transform((data) => {
      return {
        invoice: data.sequences.invoice,
        customer: data.sequences.customer,
        estimate: data.sequences.estimate,
        payment: data.sequences.payment,
        order: data.sequences.order,
        warehouse: data.sequences.warehouse,
        template: data.sequences.template,
      };
    });
    put(`/settings/${auth.account.uuid}/companies/${uuid}/sequences`, { ...headers, preserveState: 'errors' });
  };

  return (
    <div className="max-w-4xl space-y-2 py-4 **:data-[slot='card']:gap-1">
      {Object.entries(data.sequences).map(([module, value]) => (
        <Card key={module}>
          <CardHeader className="font-bold capitalize">{t(`global.${module}`)}</CardHeader>
          {errors && <GroupedValidationList errors={errors} module={module} />}
          <CardContent className="space-y-2">
            {typeof value === 'object' && 'prefix' in value ? (
              <SequenceRow t={t} module={module} type={null} config={value as SequenceConfig} onChange={updateField} />
            ) : (
              Object.entries(value as Record<string, SequenceConfig>).map(([sub, cfg]) => (
                <SequenceRow t={t} key={sub} module={module} type={sub} config={cfg} onChange={updateField} />
              ))
            )}
          </CardContent>
        </Card>
      ))}

      <Button onClick={handleSave} type="submit" disabled={processing} className="mt-1 py-2 md:text-xl">
        {processing ? (
          <>
            <Spinner />
            {t('global.saving')}
          </>
        ) : (
          <>{t('global.save')}</>
        )}
      </Button>
    </div>
  );
}

type RowProps = {
  module: string;
  type: string | null;
  config: SequenceConfig;
  onChange: (module: string, type: string | null, field: keyof SequenceConfig, value: string | number) => void;
  t: (key: string, replacements?: Replacements) => string;
};

function SequenceRow({ module, type, config, onChange, t }: RowProps) {
  const preview = `${config.prefix}${config.next.toString().padStart(config.padding, '0')}`;

  return (
    <div className="grid grid-cols-6 items-end gap-4">
      <Label className="col-span-1 font-medium capitalize">{t(`global.${type ?? module}`)}</Label>
      <Input value={config.prefix} onChange={(e) => onChange(module, type, 'prefix', e.target.value)} placeholder="Prefix" />
      <Input type="number" value={config.next} onChange={(e) => onChange(module, type, 'next', parseInt(e.target.value))} placeholder="Next" />
      <Input
        type="number"
        value={config.padding}
        onChange={(e) => onChange(module, type, 'padding', parseInt(e.target.value))}
        placeholder="Padding"
      />
      <div className="text-muted-foreground col-span-2 text-sm">{preview}</div>
    </div>
  );
}

type ErrorPayload = Record<string, string>;

interface GroupedValidationListProps {
  errors: ErrorPayload;
  module: string;
}

const extractModuleErrors = (jsonStr: Record<string, string>, module: string): Record<string, string[]> => {
  if (!jsonStr) return {};

  let parsed: Record<string, string[]>;
  try {
    parsed = JSON.parse(JSON.stringify(jsonStr));
  } catch (err) {
    console.error('Invalid JSON in sequences:', err);
    return {};
  }

  const filtered: Record<string, string[]> = {};

  for (const key in parsed) {
    if (key.startsWith(`${module}.`)) {
      const trimmedKey = key.slice(module.length + 1); // remove module and dot
      filtered[trimmedKey] = parsed[key];
    }
  }
  return filtered;
};
const GroupedValidationList: React.FC<GroupedValidationListProps> = ({ errors, module }) => {
  const moduleErrors = extractModuleErrors(errors, module);

  if (Object.keys(moduleErrors).length === 0) return null;

  return (
    <div className="mx-6.5">
      <ul className="list-inside list-disc space-y-2 text-sm text-gray-700">
        {Object.entries(moduleErrors).map(([field, messages]) => (
          <li key={field}>
            <span className="font-medium">{field}:</span> {messages.join(' | ')}
          </li>
        ))}
      </ul>
    </div>
  );
};
