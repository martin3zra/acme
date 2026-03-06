import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Pencil, Trash2 } from 'lucide-react';
import { AttributeValue } from '../../types';

type ListProps = {
  data: AttributeValue[];
  onEdit: (value: AttributeValue) => void;
  onDelete: (id: number) => void;
  t: (key: string) => string;
};

export function List({ data, onEdit, onDelete, t }: ListProps) {
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('global.displayName')}</TableHead>
            <TableHead>{t('global.code')}</TableHead>
            <TableHead>{t('attributes.values.sortOrder')}</TableHead>
            <TableHead className="text-right">{t('global.actions.title')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((value) => (
            <TableRow key={value.id}>
              <TableCell className="font-medium">{value.display_name}</TableCell>
              <TableCell>{value.value}</TableCell>
              <TableCell>{value.sort_order}</TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <button onClick={() => onEdit(value)} className="rounded p-2 hover:bg-gray-100">
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button onClick={() => onDelete(value.id)} className="rounded p-2 text-red-600 hover:bg-gray-100">
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
