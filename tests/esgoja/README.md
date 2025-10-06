# esgoja minimal test server

- Port: 8802
- Routes:
  - `/health` -> JSON ok
  - `/` -> renders `pages/index.html` (invoked as `index.html`)
  - `/test` -> renders `pages/test.html` (invoked as `test.html`)
- Static: `/static/` serves `./static`

## Template package (tpl)
- Engine: `tpl.NewEngine(".")`
- Loader: `tpl.LoadAll(engine)` and `tpl.DetectChanges(engine)`
- Transform: `tpl.TransformHTMLToTSX(html, kind)` with parser-based extraction
- Transpile: `tpl.EsbuildTranspile(tsx, baseDir)` (scaffolded)
- SSR Render: `tpl.RenderJSToString(js, props, state)` (scaffolded)

## Behavior
- Top-level names (e.g., `index.html`) resolve under `pages/`.
- If `generated/pages/<name>.shtml` exists, it is served; otherwise the source HTML is served.
- `tpl.LoadAll` processes in order: components -> layouts -> pages, writes `generated/<type>/*.tsx`.

## HTML -> TSX transform
- `tpl.TransformHTMLToTSX(html, kind)`:
  - `kind`: "component" | "layout" | "page"
  - Extracts inline `<script>` (and `<style>` kept for future), injects script above return
  - Replaces `class` with `className`
  - Converts `{{Key}}` to `{props.Key}` (placeholder for future data)
  - Names: Component -> `export default function Component`; Layout -> `Layout`; Page -> `Page`
  - Wraps body content with fragment or custom `jsxFragmentFactory` tag
  - Adds JSX pragmas and import from `jsxImportSource` (preact by default)

## AI File Status (Finalized vs UnFinalized)
- [x] tests/esgoja/main.go
- [x] tests/esgoja/tpl/engine.go
- [x] tests/esgoja/tpl/loader.go
- [x] tests/esgoja/tpl/transform.go
- [x] tests/esgoja/tpl/transform_extract.go
- [ ] tests/esgoja/tpl/transpile.go

(When you mark a file as Finalized, I’ll avoid adding new functions in it and place new ones in a new file, e.g., `transform_creativename.go`. If you change a file’s header to Finalized/UnFinalized, I’ll update this checklist accordingly.)

## Make usage (background by default)
- Start: `cd tests/esgoja && make test &`
- Stop: `cd tests/esgoja && make stop`
- Tail logs: `cd tests/esgoja && make tail`
- Vet (tidy + vet): `cd tests/esgoja && make vet`

## AI WORKFLOW RULES
Before making any changes to Finalized for AI files, ask me to proceed to change those files with the interface methods you will add. show me the modified interface on chat. list those functions with code here on chat. No private non-interface functions allowed. and once i give a green flag then change it to unfinalized for AI and then add it to the implementation. once refactored and finalized, mark it as finalized.

## HYDRATION ISSUES
A default rendering pattern if no state or default state is assumed for SSR and initialState of client rendering, with a placeholder of animated/repeating svg bg.
