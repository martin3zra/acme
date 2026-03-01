export const parseMoney = (value: string): number => {
  const clean = value.replace(/[^\d]/g, '')
  return clean ? parseInt(clean, 10) : 0
}

export const formatMoney = (cents: number) =>
  new Intl.NumberFormat('es-DO', {
    style: 'currency',
    currency: 'DOP',
  }).format(cents / 100)

  // For when we start using cents
  // export const formatMoney = (cents: number) =>
  // new Intl.NumberFormat('en-US', {
  //   style: 'currency',
  //   currency: 'USD',
  // }).format(cents / 100)