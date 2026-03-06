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
import { Plus, Pencil, Trash2 } from 'lucide-react';
import { useState } from 'react';

interface Warehouse {
  id: number;
  uuid: string;
  code: string;
  name: string;
  address?: string;
  description?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export default function Index({ warehouses }: PageProps<{ warehouses: Warehouse[] }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    address: '',
    description: '',
  });
  const [editingId, setEditingId] = useState<number | null>(null);

  const handleCreate = () => {
    setFormData({ name: '', address: '', description: '' });
    setEditingId(null);
    setOpen(true);
  };

  const handleEdit = (warehouse: Warehouse) => {
    setFormData({
      name: warehouse.name,
      address: warehouse.address || '',
      description: warehouse.description || '',
    });
    setEditingId(warehouse.id);
    setOpen(true);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (editingId) {
      router.put(`/warehouses/${editingId}`, formData);
    } else {
      router.post('/warehouses', formData);
    }

    setOpen(false);
  };

  const handleDelete = (id: number) => {
    if (confirm('Are you sure?')) {
      router.delete(`/warehouses/${id}`);
    }
  };

  const handleStatusToggle = (id: number) => {
    router.put(`/warehouses/${id}/change-status`, {});
  };

  return (
    <AppLayout
      title={t('@global.warehouses')}
      breadcrumbs={[{ label: t('@global.warehouses'), href: '/warehouses' }]}
    >
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold">{t('@global.warehouses')}</h1>
          <Button onClick={handleCreate}>
            <Plus className="w-4 h-4 mr-2" />
            {t('@global.create')}
          </Button>
        </div>

        {warehouses.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">{t('@global.noDataAvailable')}</p>
          </div>
        ) : (
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('@global.code')}</TableHead>
                  <TableHead>{t('@global.name')}</TableHead>
                  <TableHead>{t('@global.address')}</TableHead>
                  <TableHead>{t('@global.status')}</TableHead>
                  <TableHead className="text-right">{t('@global.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {warehouses.map((warehouse) => (
                  <TableRow key={warehouse.id}>
                    <TableCell className="font-medium">{warehouse.code}</TableCell>
                    <TableCell>{warehouse.name}</TableCell>
                    <TableCell>{warehouse.address || '-'}</TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <button
                          onClick={() => handleStatusToggle(warehouse.id)}
                          className={`px-3 py-1 rounded text-sm font-medium ${
                            warehouse.status === 'enabled'
                              ? 'bg-green-100 text-green-800'
                              : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {warehouse.status}
                        </button>
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => handleEdit(warehouse)}
                          className="p-2 hover:bg-gray-100 rounded"
                        >
                          <Pencil className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDelete(warehouse.id)}
                          className="p-2 hover:bg-gray-100 rounded text-red-600"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>
              {editingId ? t('@global.edit') : t('@global.create')} {t('@global.warehouse')}
            </SheetTitle>
            <SheetDescription>
              {editingId
                ? `${t('@global.update')} warehouse details`
                : `${t('@global.create')} a new warehouse`}
            </SheetDescription>
          </SheetHeader>

          <form onSubmit={handleSubmit} className="space-y-6 mt-6">
            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.name')}</label>
              <Input
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.address')}</label>
              <Input
                value={formData.address}
                onChange={(e) => setFormData({ ...formData, address: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.description')}</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full p-2 border rounded-md"
                rows={3}
              />
            </div>

            <Button type="submit" className="w-full">
              {t('@global.save')}
            </Button>
          </form>
        </SheetContent>
      </Sheet>
    </AppLayout>
  );
}
