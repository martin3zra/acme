import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { router } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { useState } from 'react';

interface StockLevel {
  id: number;
  uuid: string;
  warehouse_id: number;
  variant_id: number;
  quantity: number;
  reorder_level: number;
  reorder_quantity: number;
  created_at: string;
  updated_at: string;
}

interface Warehouse {
  id: number;
  code: string;
  name: string;
}

export default function Index({
  stocks,
  warehouses,
}: PageProps<{
  stocks: StockLevel[];
  warehouses: Warehouse[];
}>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({
    warehouse_id: '',
    variant_id: '',
    quantity: '',
    reason: '',
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    router.post('/stock-levels/adjust', formData);
    setOpen(false);
  };

  return (
    <AppLayout
      title={t('@global.stock')}
      breadcrumbs={[{ label: t('@global.stock'), href: '/stock-levels' }]}
    >
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold">{t('@global.stockLevels')}</h1>
          <Button onClick={() => setOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            {t('@global.adjustStock')}
          </Button>
        </div>

        {stocks.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">{t('@global.noDataAvailable')}</p>
          </div>
        ) : (
          <div className="border rounded-lg overflow-x-auto">
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
                {stocks.map((stock) => {
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
                          className={`inline-block px-2 py-1 rounded text-xs font-medium ${
                            isLow
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-green-100 text-green-800'
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
        )}
      </div>

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>{t('@global.adjustStock')}</SheetTitle>
            <SheetDescription>
              {t('@global.adjustStockQuantity')} for a warehouse and variant
            </SheetDescription>
          </SheetHeader>

          <form onSubmit={handleSubmit} className="space-y-6 mt-6">
            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.warehouse')}</label>
              <select
                value={formData.warehouse_id}
                onChange={(e) => setFormData({ ...formData, warehouse_id: e.target.value })}
                className="w-full p-2 border rounded-md"
                required
              >
                <option value="">Select warehouse</option>
                {warehouses.map((w) => (
                  <option key={w.id} value={w.id}>
                    {w.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.variant')}</label>
              <Input
                type="number"
                value={formData.variant_id}
                onChange={(e) => setFormData({ ...formData, variant_id: e.target.value })}
                placeholder="Variant ID"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.quantity')}</label>
              <Input
                type="number"
                value={formData.quantity}
                onChange={(e) => setFormData({ ...formData, quantity: e.target.value })}
                placeholder="Enter quantity"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.reason')}</label>
              <textarea
                value={formData.reason}
                onChange={(e) => setFormData({ ...formData, reason: e.target.value })}
                className="w-full p-2 border rounded-md"
                placeholder="Optional reason for adjustment"
                rows={3}
              />
            </div>

            <Button type="submit" className="w-full">
              {t('@global.adjustStock')}
            </Button>
          </form>
        </SheetContent>
      </Sheet>
    </AppLayout>
  );
}
