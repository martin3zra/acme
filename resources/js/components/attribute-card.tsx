import ChipSelector from '@/components/chip-selector';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useEffect, useState } from 'react';

type AttributeCardProps = {
  title: string;
  values: string[];
  selectedValues?: string[];
  onChange?: (nextValues: string[]) => void;
};

export default function AttributeCard({ title, values, selectedValues, onChange }: AttributeCardProps) {
  const [availableValues, setAvailableValues] = useState<string[]>(() => Array.from(new Set(values.map((value) => value.trim()).filter(Boolean))));
  const [selected, setSelected] = useState<string[]>(selectedValues ?? []);
  const [draftValue, setDraftValue] = useState<string>('');
  const [addError, setAddError] = useState<string>('');

  useEffect(() => {
    if (selectedValues) {
      setSelected(selectedValues);
    }
  }, [selectedValues]);

  useEffect(() => {
    setAvailableValues((current) => {
      const next = [...current];

      for (const value of values) {
        const normalized = value.trim();
        if (!normalized) {
          continue;
        }

        if (!next.some((entry) => entry.toLowerCase() === normalized.toLowerCase())) {
          next.push(normalized);
        }
      }

      return next.length === current.length ? current : next;
    });
  }, [values]);

  const toggleValue = (value: string) => {
    setSelected((current) => {
      const nextValues = current.includes(value) ? current.filter((entry) => entry !== value) : [...current, value];
      onChange?.(nextValues);
      return nextValues;
    });
  };

  const addValue = () => {
    const normalized = draftValue.trim();

    if (!normalized) {
      return;
    }

    const alreadyExists = availableValues.some((value) => value.toLowerCase() === normalized.toLowerCase());
    if (alreadyExists) {
      setAddError('Value already exists.');
      return;
    }

    setAvailableValues((current) => [...current, normalized]);
    setSelected((current) => {
      const nextValues = current.includes(normalized) ? current : [...current, normalized];
      onChange?.(nextValues);
      return nextValues;
    });
    setDraftValue('');
    setAddError('');
  };

  return (
    <section className="border-border/70 bg-background rounded-xl border p-4 sm:p-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h3 className="text-sm font-semibold tracking-tight">{title}</h3>
      </div>

      <div className="border-border/70 mt-3 rounded-lg border border-dashed p-3 sm:p-4">
        <div className="flex flex-wrap items-center gap-2">
          <ChipSelector
            className="contents"
            options={availableValues.map((value) => ({
              label: value,
              value,
            }))}
            selectedValues={selected}
            onToggle={toggleValue}
          />

          <Input
            value={draftValue}
            onChange={(event) => {
              setDraftValue(event.target.value);
              if (addError) {
                setAddError('');
              }
            }}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                event.preventDefault();
                addValue();
              }
            }}
            placeholder="Add value"
            className="h-8 w-32"
          />
          <Button type="button" size="sm" variant="outline" className="h-8" onClick={addValue}>
            Add
          </Button>
        </div>

        {availableValues.length === 0 && <p className="text-muted-foreground mt-2 text-sm">No values yet.</p>}
        {addError && <p className="mt-2 text-xs text-red-600">{addError}</p>}
      </div>
    </section>
  );
}
