function main() {
	marked.setOptions({gfm:true});
	window.contents = document.getElementById("realm_render").innerHTML;
	var doc = new DOMParser().parseFromString(window.contents, "text/html");
	var contents = doc.documentElement.textContent
	var parsed = marked.parse(contents);
	document.getElementById("realm_render").innerHTML = parsed;
};
