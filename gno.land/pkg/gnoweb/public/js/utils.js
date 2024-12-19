function r(t,n=250){let e;return function(...i){e!==void 0&&clearTimeout(e),e=setTimeout(()=>{t.apply(this,i)},n)}}export{r as debounce};
