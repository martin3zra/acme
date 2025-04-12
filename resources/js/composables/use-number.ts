export function useNumber() {
    return {currency}
}

// Set inCent to TRUE once the database structure is modified
function currency(value: number|string, precision: number = 2, inCent: boolean = false): string {

    const formatter = new Intl.NumberFormat("en-US", {
        minimumFractionDigits: precision,
        maximumFractionDigits: precision,
    })

    if (typeof value === 'string') {
        return `$${formatter.format(inCent ? parseInt(value)/100 : Number(value))}`
    }

    return `$${formatter.format(inCent ? value/100: value)}`
}