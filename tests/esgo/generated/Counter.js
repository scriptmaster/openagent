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
    default: () => Counter
  });
  function Counter(props, state) {
    let count = props && props.count != null ? props.count : state && state.count || 0;
    function increment() {
      count = count + 1;
    }
    function decrement() {
      count = count - 1;
    }
    return /* @__PURE__ */ React.createElement("div", { className: "counter" }, /* @__PURE__ */ React.createElement("button", { onClick: decrement }, "-"), /* @__PURE__ */ React.createElement("span", null, count), /* @__PURE__ */ React.createElement("button", { onClick: increment }, "+"));
  }
  return __toCommonJS(stdin_exports);
})();
