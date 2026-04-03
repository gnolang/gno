function T(l){let e=[],c=0,o=l.querySelector('[data-playground-target="code"]'),i=l.querySelector('[data-playground-target="output"]'),s=l.querySelector('[data-playground-target="tabs"]');if(!o||!i||!s)return;let m=l.getAttribute("data-playground-domain-value")||"gno.land",d=o.value;if(d.includes("// --- ")&&d.includes(" ---")){let t=d.split(/^\/\/ --- (.+?) ---$/m);for(let n=1;n<t.length;n+=2){let a=t[n].trim(),r=(t[n+1]||"").trim();a&&e.push({name:a,content:r})}e.length===0&&(e=[{name:"main.gno",content:d}]),o.value=e[0].content}else e=[{name:"main.gno",content:d}];function g(){for(;s.firstChild;)s.removeChild(s.firstChild);e.forEach((n,a)=>{let r=document.createElement("button");r.className=`b-playground-tab${a===c?" b-playground-tab--active":""}`,r.textContent=n.name,r.addEventListener("click",()=>v(n.name)),s.appendChild(r)});let t=document.createElement("button");t.className="b-playground-tab-add",t.textContent="+",t.title="Add file",t.addEventListener("click",h),s.appendChild(t)}function v(t){e[c].content=o.value;let n=e.findIndex(a=>a.name===t);n>=0&&(c=n,o.value=e[n].content,g())}function h(){let t=prompt("File name (e.g. helper.gno):");!t||!t.endsWith(".gno")||e.some(n=>n.name===t)||(e[c].content=o.value,e.push({name:t,content:`package main
`}),c=e.length-1,o.value=e[c].content,g())}async function p(){e[c].content=o.value,i.textContent="Running...",i.classList.remove("u-color-danger");let t=o.value,n=t.match(/^package\s+(\w+)/m),a=n?n[1]:"main";if(t.includes("func Render("))try{let u=await(await fetch("/_/api/eval",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({pkg_path:`${m}/r/playground_preview`,expression:'Render("")'})})).json();u.error?(i.textContent=`Error: ${u.error}`,i.classList.add("u-color-danger")):i.textContent=u.result}catch{i.textContent=`Note: Server-side execution not available for scratch pad code.

Package: ${a}
Files: ${e.map(r=>r.name).join(", ")}

To deploy and test:
  gnokey maketx addpkg -pkgpath "${m}/r/yourname/pkg" ...`}else i.textContent=`Package: ${a}
Files: ${e.map(r=>r.name).join(", ")}

To run locally:
  gno run ${e.map(r=>r.name).join(" ")}

To test:
  gno test .`}function y(){i.textContent=`Testing requires a running gno node.

To test locally:
  gno test .`}function C(){i.textContent=`Formatting requires server-side gno fmt (coming soon).

To format locally:
  gno fmt -w `+e[c].name}function b(){e[c].content=o.value;let t=e.length===1?e[0].content:e.map(a=>`// --- ${a.name} ---
${a.content}`).join(`

`),n=`${window.location.origin}/_/play?code=${encodeURIComponent(t)}`;navigator.clipboard.writeText(n).then(()=>{i.textContent="Share URL copied to clipboard!"}).catch(()=>{i.textContent=`Share URL:
${n}`})}function x(){i.textContent="// Run code to see output here",i.classList.remove("u-color-danger")}o.addEventListener("keydown",t=>{if(t.ctrlKey&&t.key==="Enter"){t.preventDefault(),p();return}if(t.key==="Tab"&&!t.shiftKey){t.preventDefault();let n=o.selectionStart,a=o.selectionEnd;o.value=o.value.substring(0,n)+"	"+o.value.substring(a),o.selectionStart=o.selectionEnd=n+1}});let E={runCode:p,runTests:y,formatCode:C,shareCode:b,clearOutput:x};l.querySelectorAll("[data-action]").forEach(t=>{let n=t.getAttribute("data-action");if(!n)return;let a=n.match(/^(\w+)->playground#(\w+)$/);if(!a)return;let[,r,u]=a,f=E[u];f&&t.addEventListener(r,f)}),g()}export{T as default};
