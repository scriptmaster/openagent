export default function PageIndex(props:any,state:any){
let title = (props && props.title) || 'Welcome';
let subtitle = (props && props.subtitle) || 'Rendered via ESBuild + Goja';
  return React.createElement(Layout, props, (
    <h1>{title}</h1>
<p>{subtitle}</p>
<Counter />
  ));
}
