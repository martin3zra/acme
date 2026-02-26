import { Replacements } from '@/types';
import { CsvPreview } from '.';
import { Button } from '../ui/button';

type Props = {
  totalRows: number;
  preview: CsvPreview | null;
  encoding: string | null;
  handleOnClear: () => void;
  t: (key: string, replacements?: Replacements) => string;
};

export function DropZonePreview({ totalRows, preview, encoding, handleOnClear, t }: Props) {
  if (preview === null) {
    return <></>;
  }
  return (
    <div className="space-y-2">
      {preview.headers.length > 0 && (
        <div className="space-y-2">
          <div className="flex w-full justify-between">
            <p className="text-sm font-medium">{t('global.import.preview')}</p>
            <Button variant={'outline'} onClick={handleOnClear}>
              {t('global.import.clear')}
            </Button>
          </div>

          <div className="max-h-64 overflow-auto rounded border">
            <table className="w-full border-collapse text-xs">
              <thead className="bg-muted sticky top-0">
                <tr>
                  {preview.headers.map((header, i) => (
                    <th key={i} className="border-b px-3 py-2 text-left font-medium">
                      {header}
                    </th>
                  ))}
                </tr>
              </thead>

              <tbody>
                {preview.rows.map((row, i) => (
                  <tr key={i} className="odd:bg-background even:bg-muted/40">
                    {row.map((cell, j) => (
                      <td key={j} className="border-b px-3 py-2 whitespace-nowrap">
                        {cell || <span className="text-muted-foreground">—</span>}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="text-xs text-green-600">{t('global.import.headerRowsDetected', { columns: preview.rows.length })}</p>

          <p className="text-muted-foreground text-xs">{t('global.import.showingFirstOf', { start: preview.rows.length, total: totalRows })}</p>
          {encoding && (
            <p className="text-muted-foreground text-xs">
              {t('global.import.encodingDetected')} <strong>{encoding}</strong>
              {encoding !== 'UTF-8' && ' (converted to UTF-8)'}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
