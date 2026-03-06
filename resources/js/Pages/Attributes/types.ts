export interface Attribute {
  id: number;
  uuid: string;
  name: string;
  type: string;
  display_name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export type AttributeForm = {
  name: string;
  type: string;
  display_name: string;
  description: string;
};

  export interface AttributeValue {
    id: number;
    uuid: string;
    attribute_id: number;
    value: string;
    display_name: string;
    sort_order: number;
    created_at: string;
    updated_at: string;
  }

  export type AttributeValueForm = {
    attribute_id: number;
    value: string;
    display_name: string;
    sort_order: number;
  };
