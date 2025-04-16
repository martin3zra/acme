export const useLocalStorage = (key: string) => {
    const setItem = (value: unknown) => {
        try {
            window.localStorage.setItem(key, JSON.stringify(value));
        } catch (e) {
            console.error(`Error setting item in localStorage: ${e}`);
        }
    }

    const getItem = () => {
        try {
            const item = window.localStorage.getItem(key);
            return item ? JSON.parse(item) : undefined;
        } catch (e) {
            console.error(`Error getting item from localStorage: ${e}`);
            return undefined;
        }
    }

    const removeItem = () => {
        try {
            window.localStorage.removeItem(key);
        } catch (e) {
            console.error(`Error removing item from localStorage: ${e}`);
        }
    }

    return { setItem, getItem, removeItem }
}