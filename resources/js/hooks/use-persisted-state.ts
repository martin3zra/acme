import { useEffect, useState } from 'react';
import { useLocalStorage } from './use-local-storage';

export function usePersistedState<T>(key: string, initialValue: T, forceInitial: boolean = false) {
  const { getItem, setItem, removeItem: deleteItem } = useLocalStorage(key);
  const [value, setValue] = useState<T>(() => {
    if (forceInitial) return initialValue;
    const item = getItem();
    return (item as T) ?? initialValue;
  });

  const removeItem = () => {
    deleteItem();
    setValue(undefined as unknown as T);
  };

  useEffect(() => {
    if (value === undefined || value === null) {
      deleteItem();
    } else {
      setItem(value);
    }
  }, [value, setItem]);

  return [value, setValue, removeItem] as const;
}
