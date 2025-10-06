/** @jsx h */
/** @jsxFrag Fragment */
import { h, Fragment, renderToString } from "../lib/ssr.js";
/* parser: golang.org/x/net/html */
export default function Page(props:any, state:any){
  return (
    <Fragment>
      <h1>{props.Title}</h1>
<p>{props.Message}</p>
<p>Static assets served from /static/</p>
    </Fragment>
  );
}
globalThis.RenderToString = function(props, state){ return renderToString(Page(props, state)); };
