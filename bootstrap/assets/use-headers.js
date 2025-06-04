import { usePage } from "@inertiajs/react";
function useHeader() {
  const { csrf_token } = usePage().props;
  return { headers: { headers: { "X-CSRF-Token": csrf_token } } };
}
export {
  useHeader as u
};
