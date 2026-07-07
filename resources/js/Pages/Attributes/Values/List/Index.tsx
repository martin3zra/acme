import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Verb } from '@/types';
import { Pencil, Trash2 } from 'lucide-react';
import { AttributeValue } from '../../types';

type ListProps = {
  data: AttributeValue[];
  onSelectValue: (value: AttributeValue, action: Verb) => void;
  t: (key: string) => string;
};

export function List({ data, onSelectValue, t }: ListProps) {
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
                  <button onClick={() => onSelectValue(value, 'edit')} className="rounded p-2 hover:bg-gray-100" title={t('global.actions.edit')}>
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onSelectValue(value, 'trash')}
                    className="rounded p-2 text-red-600 hover:bg-gray-100"
                    title={t('global.actions.delete')}
                  >
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
