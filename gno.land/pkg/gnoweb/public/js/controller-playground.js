var T="gnomod.toml";function $(d){let e=[],s=0,o=d.querySelector('[data-playground-target="code"]'),i=d.querySelector('[data-playground-target="output"]'),l=d.querySelector('[data-playground-target="tabs"]');if(!o||!i||!l)return;let m=d.getAttribute("data-playground-domain-value")||"gno.land",u=o.value;if(u.includes("// --- ")&&u.includes(" ---")){let t=u.split(/^\/\/ --- (.+?) ---$/m);for(let n=1;n<t.length;n+=2){let a=t[n].trim(),r=(t[n+1]||"").trim();a&&e.push({name:a,content:r})}e.length===0&&(e=[{name:"main.gno",content:u}]),o.value=e[0].content}else e=[{name:"main.gno",content:u}];function g(){for(;l.firstChild;)l.removeChild(l.firstChild);e.forEach((n,a)=>{let r=document.createElement("button");r.className=`b-playground-tab${a===s?" b-playground-tab--active":""}`,r.textContent=n.name,r.addEventListener("click",()=>v(n.name)),l.appendChild(r)});let t=document.createElement("button");t.className="b-playground-tab-add",t.textContent="+",t.title="Add file",t.addEventListener("click",h),l.appendChild(t)}function v(t){e[s].content=o.value;let n=e.findIndex(a=>a.name===t);n>=0&&(s=n,o.value=e[n].content,g())}function h(){let t=prompt("File name (e.g. helper.gno):"),n=t===T;!t||!(t.endsWith(".gno")||n)||e.some(a=>a.name===t)||(e[s].content=o.value,e.push({name:t,content:n?"":`package main
`}),s=e.length-1,o.value=e[s].content,g())}async function p(){e[s].content=o.value,i.textContent="Running...",i.classList.remove("u-color-danger");let t=o.value,n=t.match(/^package\s+(\w+)/m),a=n?n[1]:"main";if(t.includes("func Render("))try{let c=await(await fetch("/_/api/eval",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({pkg_path:`${m}/r/playground_preview`,expression:'Render("")'})})).json();c.error?(i.textContent=`Error: ${c.error}`,i.classList.add("u-color-danger")):i.textContent=c.result}catch{i.textContent=`Note: Server-side execution not available for scratch pad code.

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
  gno fmt -w `+e[s].name}function b(){e[s].content=o.value;let t=e.length===1?e[0].content:e.map(c=>`// --- ${c.name} ---
${c.content}`).join(`

`),n=new TextEncoder().encode(t),a=Array.from(n,c=>String.fromCharCode(c)).join(""),r=`${window.location.origin}/_/play?code=${encodeURIComponent(btoa(a))}`;navigator.clipboard.writeText(r).then(()=>{i.textContent="Share URL copied to clipboard!"}).catch(()=>{i.textContent=`Share URL:
${r}`})}function E(){i.textContent="// Run code to see output here",i.classList.remove("u-color-danger")}o.addEventListener("keydown",t=>{if(t.ctrlKey&&t.key==="Enter"){t.preventDefault(),p();return}if(t.key==="Tab"&&!t.shiftKey){t.preventDefault();let n=o.selectionStart,a=o.selectionEnd;o.value=`${o.value.substring(0,n)}	${o.value.substring(a)}`,o.selectionStart=o.selectionEnd=n+1}});let x={runCode:p,runTests:y,formatCode:C,shareCode:b,clearOutput:E};d.querySelectorAll("[data-action]").forEach(t=>{let n=t.getAttribute("data-action");if(!n)return;let a=n.match(/^(\w+)->playground#(\w+)$/);if(!a)return;let[,r,c]=a,f=x[c];f&&t.addEventListener(r,f)}),g()}export{$ as default};
