export {};

declare global {
  interface Window {
    _abilities?: Record<string, boolean>; // Replace with more accurate type if needed
  }
}
