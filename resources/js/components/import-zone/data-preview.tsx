import { CsvPreview } from '.';
import { Button } from '../ui/button';

type Props = {
  totalRows: number;
  preview: CsvPreview | null;
  encoding: string | null;
  handleOnClear: () => void;
};

export function DropZonePreview({ totalRows, preview, encoding, handleOnClear }: Props) {
  if (preview === null) {
    return <></>;
  }
  return (
    <div className="space-y-2">
      {preview.headers.length > 0 && (
        <div className="space-y-2">
          <div className="flex w-full justify-between">
            <p className="text-sm font-medium">Preview</p>
            <Button variant={'outline'} onClick={handleOnClear}>
              Clear
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
          <p className="text-xs text-green-600">Header row detected • {preview.headers.length} columns found</p>

          <p className="text-muted-foreground text-xs">
            Showing first {preview.rows.length} of {totalRows} rows
          </p>
          {encoding && (
            <p className="text-muted-foreground text-xs">
              Encoding detected: <strong>{encoding}</strong>
              {encoding !== 'UTF-8' && ' (converted to UTF-8)'}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
