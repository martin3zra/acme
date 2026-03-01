import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

interface ImportError {
  row: number;
  error: string;
  raw: string;
}

export function ImportErrorTable({ errors }: { errors: ImportError[] }) {
  if (!errors.length) return null;

  return (
    <div className="mt-4">
      <h4 className="mb-2 font-medium">Skipped rows</h4>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Row</TableHead>
            <TableHead>Error</TableHead>
            <TableHead>Data</TableHead>
          </TableRow>
        </TableHeader>

        <TableBody>
          {errors.map((e, i) => (
            <TableRow key={i}>
              <TableCell>{e.row}</TableCell>
              <TableCell className="text-red-600">{e.error}</TableCell>
              <TableCell className="max-w-75 truncate">{e.raw}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
