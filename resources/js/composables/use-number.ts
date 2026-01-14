export function useNumber() {
  return { currency, abbrvNumber };
}

// Set inCent to TRUE once the database structure is modified
function currency(value: number | string, precision: number = 2, inCent: boolean = false): string {
  const formatter = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: precision,
    maximumFractionDigits: precision,
  });

  if (typeof value === 'string') {
    return `$${formatter.format(inCent ? parseInt(value) / 100 : Number(value))}`;
  }

  return `$${formatter.format(inCent ? value / 100 : value)}`;
}

export function abbrvNumber(value: number): string {
  if (value === null || value === undefined) return '';
  if (value < 1000) return value.toString();

  const units = ['K', 'M', 'B', 'T']; // thousand, million, billion, trillion
  const order = Math.floor(Math.log10(value) / 3);
  const unitname = units[order - 1];
  const num = value / Math.pow(1000, order);

  // Keep one decimal if not an integer
  return num % 1 === 0 ? num.toString() + unitname : num.toFixed(1) + unitname;
}
