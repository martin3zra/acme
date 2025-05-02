export function useDocument() {
  return { removeElementParent };
}

function removeElementParent(e: Event, tag: string) {
  const modal = document.getElementsByTagName(tag)[0];
  if (modal?.parentNode) {
    const grantFather = modal.parentNode?.parentNode;
    grantFather?.removeChild(modal.parentNode);
    e.preventDefault();
  }
}
