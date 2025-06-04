function useVerb() {
  return { action };
}
function action(verb) {
  return {
    view: "view",
    edit: "update",
    trash: "trash",
    create: "create"
  }[verb];
}
export {
  useVerb as u
};
