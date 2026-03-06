import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Pencil, Trash2 } from 'lucide-react';
import { Warehouse } from '../types';

type ListProps = {
  data: Warehouse[];
  onEdit: (warehouse: Warehouse) => void;
  onDelete: (id: number) => void;
  onStatusToggle: (id: number) => void;
  t: (key: string) => string;
};

export function List({ data, onEdit, onDelete, onStatusToggle, t }: ListProps) {
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('global.code')}</TableHead>
            <TableHead>{t('global.name')}</TableHead>
            <TableHead>{t('global.address')}</TableHead>
            <TableHead>{t('global.status')}</TableHead>
            <TableHead className="text-right">{t('global.actions.title')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((warehouse) => (
            <TableRow key={warehouse.id}>
              <TableCell className="font-medium">{warehouse.code}</TableCell>
              <TableCell>{warehouse.name}</TableCell>
              <TableCell>{warehouse.address || '-'}</TableCell>
              <TableCell>
                <div className="flex gap-2">
                  <button
                    onClick={() => onStatusToggle(warehouse.id)}
                    className={`rounded px-3 py-1 text-sm font-medium ${warehouse.status === 'enabled' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}
                  >
                    {warehouse.status}
                  </button>
                </div>
              </TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <button onClick={() => onEdit(warehouse)} className="rounded p-2 hover:bg-gray-100">
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button onClick={() => onDelete(warehouse.id)} className="rounded p-2 text-red-600 hover:bg-gray-100">
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
