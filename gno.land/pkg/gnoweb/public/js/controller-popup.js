import{BaseController as r}from"./controller.js";var n=class extends r{connect(){}key(e){let{key:t}=e;t!=="Enter"&&t!==" "||(e.preventDefault(),this.element.click())}};export{n as PopupController};
