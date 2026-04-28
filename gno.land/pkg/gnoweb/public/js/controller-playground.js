import{BaseController as l}from"./controller.js";var r="gnomod.toml",c=`package main
`,a=class extends l{connect(){this.files=[],this.activeFile=0,this.codeEl=this.getTarget("code"),this.outputEl=this.getTarget("output"),this.tabsEl=this.getTarget("tabs"),!(!this.codeEl||!this.outputEl||!this.tabsEl)&&(this._parseInitialCode(),this._setupKeyboardShortcuts(),this.renderTabs())}_parseInitialCode(){let t=this.codeEl.value;if(t.includes("// --- ")&&t.includes(" ---")){let e=t.split(/^\/\/ --- (.+?) ---$/m);for(let i=1;i<e.length;i+=2){let n=e[i].trim(),s=(e[i+1]||"").trim();n&&this.files.push({name:n,content:s})}this.files.length===0&&(this.files=[{name:"main.gno",content:t}]),this.codeEl.value=this.files[0].content}else this.files=[{name:"main.gno",content:t}]}_setupKeyboardShortcuts(){this.codeEl.addEventListener("keydown",t=>{if(t.ctrlKey&&t.key==="Enter"){t.preventDefault(),this.runCode();return}if(t.key==="Tab"&&!t.shiftKey){t.preventDefault();let e=this.codeEl.selectionStart,i=this.codeEl.selectionEnd;this.codeEl.value=`${this.codeEl.value.substring(0,e)}	${this.codeEl.value.substring(i)}`,this.codeEl.selectionStart=this.codeEl.selectionEnd=e+1}})}_setOutput(t,e=!1){this.outputEl.textContent=t,this.outputEl.classList.toggle("u-color-danger",e)}_switchToFile(t){this.files[this.activeFile].content=this.codeEl.value;let e=this.files.findIndex(i=>i.name===t);return e>=0&&(this.activeFile=e,this.codeEl.value=this.files[e].content,this.renderTabs()),e>=0}renderTabs(){for(;this.tabsEl.firstChild;)this.tabsEl.removeChild(this.tabsEl.firstChild);this.files.forEach((e,i)=>{let n=document.createElement("button");n.className=`b-playground-tab${i===this.activeFile?" b-playground-tab--active":""}`,n.textContent=e.name,n.addEventListener("click",()=>this._switchToFile(e.name)),this.tabsEl.appendChild(n)});let t=document.createElement("button");t.className="b-playground-tab-add",t.textContent="+",t.title="Add file",t.addEventListener("click",()=>this.addFile()),this.tabsEl.appendChild(t)}switchTab(t){let e=t.params?.file;e&&this._switchToFile(e)}addFile(){let t=prompt("File name (e.g. helper.gno):");if(t==null||this._switchToFile(t))return;let e=t===r;if(!t.endsWith(".gno")&&!e)return;let i=this.getValue("domain")||"gno.land",n=c;e&&(n=`module = "${i}/r/yourname/pkg"
gno = "0.9"`),this.files[this.activeFile].content=this.codeEl.value,this.files.push({name:t,content:n}),this.activeFile=this.files.length-1,this.codeEl.value=this.files[this.activeFile].content,this.renderTabs()}async runCode(){this.files[this.activeFile].content=this.codeEl.value,this._setOutput("Running...");let t=this.codeEl.value,e=t.match(/^package\s+(\w+)/m),i=e?e[1]:"main",n=this.getValue("domain")||"gno.land";if(t.includes("func Render("))try{let o=await(await fetch("/_/api/eval",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({pkg_path:`${n}/r/playground_preview`,expression:'Render("")'})})).json();o.error?this._setOutput(`Error: ${o.error}`,!0):this._setOutput(o.result)}catch{this._setOutput(`Note: Server-side execution not available for scratch pad code.

Package: ${i}
Files: ${this.files.map(s=>s.name).join(", ")}

To deploy and test:
  gnokey maketx addpkg -pkgpath "${n}/r/yourname/pkg" ...`)}else this._setOutput(`Package: ${i}
Files: ${this.files.map(s=>s.name).join(", ")}

To run locally:
  gno run ${this.files.map(s=>s.name).join(" ")}

To test:
  gno test .`)}runTests(){this._setOutput(`Testing requires a running gno node.

To test locally:
  gno test .`)}formatCode(){this._setOutput(`Formatting requires server-side gno fmt (coming soon).

To format locally:
  gno fmt -w `+this.files[this.activeFile].name)}shareCode(){this.files[this.activeFile].content=this.codeEl.value;let t=this.files.length===1?this.files[0].content:this.files.map(s=>`// --- ${s.name} ---
${s.content}`).join(`

`),e=new TextEncoder().encode(t),i=Array.from(e,s=>String.fromCharCode(s)).join(""),n=`${window.location.origin}/_/play?code=${encodeURIComponent(btoa(i))}`;navigator.clipboard.writeText(n).then(()=>{this._setOutput("Share URL copied to clipboard!")}).catch(()=>{this._setOutput(`Share URL:
${n}`)})}clearOutput(){this._setOutput("// Run code to see output here")}};export{a as PlaygroundController};
