function d(n,r=300){let e;return(...t)=>{e!==void 0&&clearTimeout(e),e=window.setTimeout(()=>{n(...t)},r)}}export{d as debounce};
