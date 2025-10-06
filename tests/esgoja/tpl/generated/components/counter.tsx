/** @jsx h */
/** @jsxFrag Fragment */
import { h, Fragment, renderToString } from "../lib/ssr.js";
/* parser: golang.org/x/net/html */
export default function Component(props:any, state:any){
let count = (props && props.count) || 3;
function inc(){ count = count + 1 }
  return (
    <Fragment>
      <button className="btn" onclick="inc()">Count: {count}</button>
    </Fragment>
  );
}
globalThis.RenderToString = function(props, state){ return renderToString(Component(props, state)); };
