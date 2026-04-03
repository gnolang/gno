import{BaseController as r}from"./controller.js";var l=class extends r{files=[];activeFile=0;_codeEl;_outputEl;_tabsEl;connect(){this._codeEl=this.getTarget("code"),this._outputEl=this.getTarget("output"),this._tabsEl=this.getTarget("tabs");let t=this._codeEl.value;t.includes("// --- ")&&t.includes(" ---")?this._parseForkedFiles(t):this.files=[{name:"main.gno",content:t}],this._renderTabs(),this._setupKeyboard(),this._bindButtons()}_bindButtons(){this.element.querySelectorAll("[data-action]").forEach(t=>{let e=t.getAttribute("data-action");if(!e)return;let i=e.match(/^(\w+)->playground#(\w+)$/);if(!i)return;let[,n,o]=i,s=this[o];typeof s=="function"&&t.addEventListener(n,s.bind(this))})}_parseForkedFiles(t){let e=t.split(/^\/\/ --- (.+?) ---$/m);this.files=[];for(let i=1;i<e.length;i+=2){let n=e[i].trim(),o=(e[i+1]||"").trim();n&&this.files.push({name:n,content:o})}this.files.length===0&&(this.files=[{name:"main.gno",content:t}]),this._codeEl.value=this.files[0].content}_setupKeyboard(){this._codeEl.addEventListener("keydown",t=>{if(t.ctrlKey&&t.key==="Enter"){t.preventDefault(),this.runCode();return}if(t.key==="Tab"&&!t.shiftKey){t.preventDefault();let e=this._codeEl.selectionStart,i=this._codeEl.selectionEnd;this._codeEl.value=this._codeEl.value.substring(0,e)+"	"+this._codeEl.value.substring(i),this._codeEl.selectionStart=this._codeEl.selectionEnd=e+1}})}_renderTabs(){for(;this._tabsEl.firstChild;)this._tabsEl.removeChild(this._tabsEl.firstChild);this.files.forEach((e,i)=>{let n=document.createElement("button");n.className=`b-playground-tab${i===this.activeFile?" b-playground-tab--active":""}`,n.textContent=e.name,n.addEventListener("click",()=>this._switchToFile(e.name)),this._tabsEl.appendChild(n)});let t=document.createElement("button");t.className="b-playground-tab-add",t.textContent="+",t.title="Add file",t.addEventListener("click",()=>this.addFile()),this._tabsEl.appendChild(t)}_switchToFile(t){this.files[this.activeFile].content=this._codeEl.value;let e=this.files.findIndex(i=>i.name===t);e>=0&&(this.activeFile=e,this._codeEl.value=this.files[e].content,this._renderTabs())}switchTab(t){let i=t.currentTarget.dataset.playgroundFileParam||"";this._switchToFile(i)}addFile(){let t=prompt("File name (e.g. helper.gno):");if(t){if(!t.endsWith(".gno")){alert("File name must end with .gno");return}if(this.files.some(e=>e.name===t)){alert("File already exists");return}this.files[this.activeFile].content=this._codeEl.value,this.files.push({name:t,content:`package main
`}),this.activeFile=this.files.length-1,this._codeEl.value=this.files[this.activeFile].content,this._renderTabs()}}async runCode(){this.files[this.activeFile].content=this._codeEl.value,this._outputEl.textContent="Running...";let t=this.getValue("remote"),e=this.getValue("domain"),i=this._codeEl.value,n=i.match(/^package\s+(\w+)/m),o=n?n[1]:"main";if(i.includes("func Render("))try{let a=await(await fetch("/_/api/eval",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({pkg_path:`${e}/r/playground_preview`,expression:'Render("")'})})).json();a.error?(this._outputEl.textContent=`Error: ${a.error}`,this._outputEl.classList.add("u-color-danger")):(this._outputEl.textContent=a.result,this._outputEl.classList.remove("u-color-danger"))}catch{this._outputEl.textContent=`Note: Server-side execution requires a running gno node.

Package: ${o}
Files: ${this.files.map(s=>s.name).join(", ")}

To deploy and test, use:
  gnokey maketx addpkg -pkgpath "${e}/r/yourname/pkg" ...`,this._outputEl.classList.remove("u-color-danger")}else this._outputEl.textContent=`Package: ${o}
Files: ${this.files.map(s=>s.name).join(", ")}

To run locally:
  gno run ${this.files.map(s=>s.name).join(" ")}

To test:
  gno test .`,this._outputEl.classList.remove("u-color-danger")}runTests(){this._outputEl.textContent=`Testing requires a running gno node.

To test locally:
  gno test .`}formatCode(){this._outputEl.textContent=`Formatting requires server-side gno fmt (coming soon).

To format locally:
  gno fmt -w `+this.files[this.activeFile].name}shareCode(){this.files[this.activeFile].content=this._codeEl.value;let t=this.files.length===1?this.files[0].content:this.files.map(n=>`// --- ${n.name} ---
${n.content}`).join(`

`),e=encodeURIComponent(t),i=`${window.location.origin}/_/play?code=${e}`;navigator.clipboard.writeText(i).then(()=>{this._outputEl.textContent="Share URL copied to clipboard!"}).catch(()=>{this._outputEl.textContent=`Share URL:
${i}`})}clearOutput(){this._outputEl.textContent="// Run code to see output here",this._outputEl.classList.remove("u-color-danger")}};export{l as PlaygroundController};
