(() => {
  // generated/lib/ssr.js
  function h(type, props, ...children) {
    return { type, props: props || {}, children };
  }
  var Fragment = Symbol.for("ssr.fragment");
  function renderToString(node) {
    try {
      if (node == null) return "";
      if (typeof node === "string" || typeof node === "number") return String(node);
      if (Array.isArray(node)) return node.map(renderToString).join("");
      if (typeof node.type === "function") {
        const rendered = node.type({ ...node.props || {}, children: node.children });
        return renderToString(rendered);
      }
      const type = node.type === Fragment ? Fragment : node.type || "div";
      const props = node.props || {};
      const children = node.children != null ? [].concat(node.children) : [];
      if (type === Fragment) {
        return children.map(renderToString).join("");
      }
      const attrs = Object.keys(props).filter((k) => k !== "children" && props[k] != null && props[k] !== false).map((k) => " " + k + '="' + escapeHtml(String(props[k])) + '"').join("");
      return "<" + type + attrs + ">" + children.map(renderToString).join("") + "</" + type + ">";
    } catch (_) {
      return "";
    }
  }
  function escapeHtml(s) {
    return s.replace(/[&<>"]+/g, function(ch) {
      switch (ch) {
        case "&":
          return "&amp;";
        case "<":
          return "&lt;";
        case ">":
          return "&gt;";
        case '"':
          return "&quot;";
        default:
          return ch;
      }
    });
  }

  // generated/pages/input.tsx
  function Page(props, state) {
    return  h(Fragment, null,  h("h1", null, props.Title),  h("p", null, props.Message),  h("p", null, "Static assets served from /static/"));
  }
  globalThis.RenderToString = function(props, state) {
    return renderToString(Page(props, state));
  };
})();
