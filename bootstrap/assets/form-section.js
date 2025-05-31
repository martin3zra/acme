import { jsxs, jsx, Fragment } from "react/jsx-runtime";
import React__default from "react";
function Title$1({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
function Description$1({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
function Aside({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
const _SectionTitle = class _SectionTitle extends React__default.PureComponent {
  render() {
    const { children } = this.props;
    const title = React__default.Children.toArray(children).find((children2) => children2.type === Title$1);
    const description = React__default.Children.toArray(children).find((children2) => children2.type === Description$1);
    const aside = React__default.Children.toArray(children).find((children2) => children2.type === Aside);
    return /* @__PURE__ */ jsxs("div", { className: "flex justify-between md:col-span-1", children: [
      /* @__PURE__ */ jsxs("div", { className: "px-4 sm:px-0", children: [
        /* @__PURE__ */ jsx("h3", { className: "text-lg font-medium text-gray-900 dark:text-gray-100", children: title ? title : null }),
        /* @__PURE__ */ jsx("p", { className: "mt-1 text-sm text-gray-600 dark:text-gray-400", children: description ? description : null })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "px-4 sm:px-0", children: aside ? aside : null })
    ] });
  }
};
_SectionTitle.Title = Title$1;
_SectionTitle.Description = Description$1;
_SectionTitle.Aside = Aside;
let SectionTitle = _SectionTitle;
function Title({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
function Description({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
function Form({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
function Actions({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
const _FormSection = class _FormSection extends React__default.Component {
  constructor() {
    super(...arguments);
    this.onSubmit = (event) => {
      event.preventDefault();
      this.props.onSubmit();
    };
  }
  render() {
    const { children } = this.props;
    const title = React__default.Children.toArray(children).find((children2) => children2.type === Title);
    const description = React__default.Children.toArray(children).find((children2) => children2.type === Description);
    const form = React__default.Children.toArray(children).find((children2) => children2.type === Form);
    const actions = React__default.Children.toArray(children).find((children2) => children2.type === Actions);
    return /* @__PURE__ */ jsxs("div", { className: "md:grid md:grid-cols-3 md:gap-6", children: [
      /* @__PURE__ */ jsxs(SectionTitle, { children: [
        /* @__PURE__ */ jsx(SectionTitle.Title, { children: title }),
        /* @__PURE__ */ jsx(SectionTitle.Description, { children: description })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "mt-5 md:col-span-2 md:mt-0", children: /* @__PURE__ */ jsxs("form", { onSubmit: this.onSubmit, children: [
        /* @__PURE__ */ jsx("div", { className: `bg-white px-4 py-5 shadow sm:p-6 dark:bg-gray-800 ${actions ? "sm:rounded-tl-md sm:rounded-tr-md" : "sm:rounded-md"}`, children: /* @__PURE__ */ jsx("div", { className: "grid grid-cols-6 gap-6", children: form }) }),
        actions && /* @__PURE__ */ jsx("div", { className: "flex items-center justify-end bg-gray-50 px-4 py-3 text-end shadow sm:rounded-br-md sm:rounded-bl-md sm:px-6 dark:bg-gray-800", children: actions })
      ] }) })
    ] });
  }
};
_FormSection.Title = Title;
_FormSection.Description = Description;
_FormSection.Form = Form;
_FormSection.Actions = Actions;
let FormSection = _FormSection;
export {
  FormSection as F
};
