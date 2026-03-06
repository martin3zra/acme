export interface StockLevel {
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

export interface Warehouse {
  id: number;
  code: string;
  name: string;
}

export type StockAdjustmentForm = {
  warehouse_id: string;
  variant_id: string;
  quantity: string;
  reason: string;
};
