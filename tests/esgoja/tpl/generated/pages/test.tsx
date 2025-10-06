/** @jsx h */
/** @jsxFrag Fragment */
import { h, Fragment, renderToString } from "../lib/ssr.js";
/* parser: golang.org/x/net/html */
export default function Page(props:any, state:any){
let title = "Test";
    let message = "Welcome to Test page.";
  return (
    <Fragment>
      <h1>{title}</h1>
<p>{message}</p>
<div>Test route OK</div>
    </Fragment>
  );
}
globalThis.RenderToString = function(props, state){ return renderToString(Page(props, state)); };
