import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { usePage } from '@inertiajs/react';
import { useEffect, useState } from 'react';
import { DateRange } from 'react-day-picker';

export function DateRangeQuickSelect({ initialPreset, onChange }: { initialPreset: string; onChange: (range: DateRange | undefined) => void }) {
  const { props } = usePage<{ dateRanges: { label: string; key: string; from?: string; to?: string }[] }>();
  const { dateRanges } = props;

  const [selectedKey, setSelectedKey] = useState<string>(initialPreset);
  useEffect(() => {
    const preset = dateRanges.find((p) => p.key === initialPreset);
    if (preset?.from) {
      onChange({
        from: new Date(preset.from),
        to: preset.to ? new Date(preset.to) : undefined,
      });
    } else {
      onChange(undefined);
    }
  }, [initialPreset, dateRanges]);

  const handleChange = (key: string) => {
    setSelectedKey(key);
    const preset = dateRanges.find((p) => p.key === key);
    if (!preset) return;

    if (preset.key === 'custom') {
      // Let user open calendar manually
      onChange(undefined);
    } else {
      onChange(
        preset.from
          ? {
              from: new Date(preset.from),
              to: preset.to ? new Date(preset.to) : undefined,
            }
          : undefined,
      );
    }
  };

  return (
    <Select value={selectedKey} onValueChange={handleChange}>
      <SelectTrigger className="w-[280px]">
        <SelectValue placeholder="Select period" />
      </SelectTrigger>
      <SelectContent>
        {dateRanges.map((preset) => (
          <SelectItem key={preset.key} value={preset.key}>
            {preset.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
