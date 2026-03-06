import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ChevronRight, Pencil, Trash2 } from 'lucide-react';
import { Attribute } from '../types';

type ListProps = {
  data: Attribute[];
  onEdit: (attribute: Attribute) => void;
  onDelete: (id: number) => void;
  t: (key: string) => string;
};

export function List({ data, onEdit, onDelete, t }: ListProps) {
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('@global.name')}</TableHead>
            <TableHead>{t('@global.displayName')}</TableHead>
            <TableHead>{t('@global.type')}</TableHead>
            <TableHead>{t('@global.description')}</TableHead>
            <TableHead className="text-right">{t('@global.actions')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((attribute) => (
            <TableRow key={attribute.id}>
              <TableCell className="font-medium">{attribute.name}</TableCell>
              <TableCell>{attribute.display_name}</TableCell>
              <TableCell>
                <span className="inline-block rounded bg-blue-100 px-2 py-1 text-sm text-blue-800">{attribute.type}</span>
              </TableCell>
              <TableCell>{attribute.description || '-'}</TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <a href={`/attributes/${attribute.id}/values`} className="inline-block rounded p-2 hover:bg-gray-100" title="Manage values">
                    <ChevronRight className="h-4 w-4" />
                  </a>
                  <button onClick={() => onEdit(attribute)} className="rounded p-2 hover:bg-gray-100">
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button onClick={() => onDelete(attribute.id)} className="rounded p-2 text-red-600 hover:bg-gray-100">
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
