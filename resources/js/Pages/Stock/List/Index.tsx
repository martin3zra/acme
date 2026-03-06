import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { StockLevel } from '../types';

type ListProps = {
  data: StockLevel[];
  t: (key: string) => string;
};

export function List({ data, t }: ListProps) {
  return (
    <div className="overflow-x-auto rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('@global.warehouse')}</TableHead>
            <TableHead>{t('@global.variant')}</TableHead>
            <TableHead className="text-right">{t('@global.quantity')}</TableHead>
            <TableHead className="text-right">{t('@global.reorderLevel')}</TableHead>
            <TableHead className="text-right">{t('@global.reorderQuantity')}</TableHead>
            <TableHead>{t('@global.status')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((stock) => {
            const isLow = stock.quantity <= stock.reorder_level;
            return (
              <TableRow key={stock.id} className={isLow ? 'bg-yellow-50' : ''}>
                <TableCell>{stock.warehouse_id}</TableCell>
                <TableCell>{stock.variant_id}</TableCell>
                <TableCell className="text-right font-medium">{stock.quantity}</TableCell>
                <TableCell className="text-right">{stock.reorder_level}</TableCell>
                <TableCell className="text-right">{stock.reorder_quantity}</TableCell>
                <TableCell>
                  <span
                    className={`inline-block rounded px-2 py-1 text-xs font-medium ${
                      isLow ? 'bg-yellow-100 text-yellow-800' : 'bg-green-100 text-green-800'
                    }`}
                  >
                    {isLow ? 'Low Stock' : 'In Stock'}
                  </span>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
