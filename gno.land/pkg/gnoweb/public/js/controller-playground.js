var w="gnomod.toml",L=`package main
`;function x(d){let e=[],l=0,y=-1,a=d.querySelector('[data-playground-target="code"]'),g=d.querySelector('[data-playground-target="output"]'),s=d.querySelector('[data-playground-target="tabs"]');if(!a||!g||!s)return;let m=d.getAttribute("data-playground-domain-value")||"gno.land",u=a.value;if(u.includes("// --- ")&&u.includes(" ---")){let n=u.split(/^\/\/ --- (.+?) ---$/m);for(let t=1;t<n.length;t+=2){let o=n[t].trim(),r=(n[t+1]||"").trim();o&&e.push({name:o,content:r})}e.length===0&&(e=[{name:"main.gno",content:u}]),a.value=e[0].content}else e=[{name:"main.gno",content:u}];function p(){for(;s.firstChild;)s.removeChild(s.firstChild);e.forEach((t,o)=>{let r=document.createElement("button");r.className=`b-playground-tab${o===l?" b-playground-tab--active":""}`,r.textContent=t.name,r.addEventListener("click",()=>f(t.name)),s.appendChild(r)});let n=document.createElement("button");n.className="b-playground-tab-add",n.textContent="+",n.title="Add file",n.addEventListener("click",b),s.appendChild(n)}function f(n){e[l].content=a.value;let t=e.findIndex(o=>o.name===n);return t>=0&&(l=t,a.value=e[t].content,p()),t>=0}function b(){let n=prompt("File name (e.g. helper.gno):");if(n==null||f(n))return;let t=n===w;if(!n.endsWith(".gno")&&!t)return;let o=L;t&&(o=`module = "${m}/r/yourname/pkg"
gno = "0.9"`,y=e.length),e[l].content=a.value,e.push({name:n,content:o}),l=e.length-1,a.value=e[l].content,p()}function i(n,t=!1){g.textContent=n,g.classList.toggle("u-color-danger",t)}async function v(){e[l].content=a.value,i("Running...");let n=a.value,t=n.match(/^package\s+(\w+)/m),o=t?t[1]:"main";if(n.includes("func Render("))try{let c=await(await fetch("/_/api/eval",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({pkg_path:`${m}/r/playground_preview`,expression:'Render("")'})})).json();c.error?i(`Error: ${c.error}`,!0):i(c.result)}catch{i(`Note: Server-side execution not available for scratch pad code.

Package: ${o}
Files: ${e.map(r=>r.name).join(", ")}

To deploy and test:
  gnokey maketx addpkg -pkgpath "${m}/r/yourname/pkg" ...`)}else i(`Package: ${o}
Files: ${e.map(r=>r.name).join(", ")}

To run locally:
  gno run ${e.map(r=>r.name).join(" ")}

To test:
  gno test .`)}function E(){i(`Testing requires a running gno node.

To test locally:
  gno test .`)}function T(){i(`Formatting requires server-side gno fmt (coming soon).

To format locally:
  gno fmt -w `+e[l].name)}function $(){e[l].content=a.value;let n=e.length===1?e[0].content:e.map(c=>`// --- ${c.name} ---
${c.content}`).join(`

`),t=new TextEncoder().encode(n),o=Array.from(t,c=>String.fromCharCode(c)).join(""),r=`${window.location.origin}/_/play?code=${encodeURIComponent(btoa(o))}`;navigator.clipboard.writeText(r).then(()=>{i("Share URL copied to clipboard!")}).catch(()=>{i(`Share URL:
${r}`)})}function k(){i("// Run code to see output here")}a.addEventListener("keydown",n=>{if(n.ctrlKey&&n.key==="Enter"){n.preventDefault(),v();return}if(n.key==="Tab"&&!n.shiftKey){n.preventDefault();let t=a.selectionStart,o=a.selectionEnd;a.value=`${a.value.substring(0,t)}	${a.value.substring(o)}`,a.selectionStart=a.selectionEnd=t+1}});let C={runCode:v,runTests:E,formatCode:T,shareCode:$,clearOutput:k};d.querySelectorAll("[data-action]").forEach(n=>{let t=n.getAttribute("data-action");if(!t)return;let o=t.match(/^(\w+)->playground#(\w+)$/);if(!o)return;let[,r,c]=o,h=C[c];h&&n.addEventListener(r,h)}),p()}export{x as default};
