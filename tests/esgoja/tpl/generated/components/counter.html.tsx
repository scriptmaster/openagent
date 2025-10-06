/** @jsx h */
/** @jsxFrag Fragment */
import { h, Fragment } from "preact";
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
