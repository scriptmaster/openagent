var Component = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

  // <stdin>
  var stdin_exports = {};
  __export(stdin_exports, {
    default: () => PageIndex
  });
  function Layout(props, state) {
    let siteTitle = props && props.siteTitle || "Home";
    return /* @__PURE__ */ React.createElement("div", { className: "layout-index" }, /* @__PURE__ */ React.createElement("header", null, /* @__PURE__ */ React.createElement("h1", null, siteTitle)), /* @__PURE__ */ React.createElement("main", null, children), /* @__PURE__ */ React.createElement("footer", null, /* @__PURE__ */ React.createElement("small", null, "Index Footer")));
  }
  function PageIndex(props, state) {
    let title = props && props.title || "Welcome";
    let subtitle = props && props.subtitle || "Rendered via ESBuild + Goja";
    const Page = /* @__PURE__ */ React.createElement("main", { className: "home" }, /* @__PURE__ */ React.createElement("h1", null, title), /* @__PURE__ */ React.createElement("p", null, subtitle), /* @__PURE__ */ React.createElement("template", { id: "component-counter" }), /* @__PURE__ */ React.createElement("script", { src: "/static/common_client.js" }), /* @__PURE__ */ React.createElement("script", { src: "/static/pages_index.js" }));
    return React.createElement(Layout, props, Page);
  }
  return __toCommonJS(stdin_exports);
})();
