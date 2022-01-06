function main() {
	marked.setOptions({gfm:true});
	window.contents = document.getElementById("contents").innerHTML;
	console.log(window.contents);
	var doc = new DOMParser().parseFromString(window.contents, "text/html");
	var contents = doc.documentElement.textContent
	console.log(contents);
	var parsed = marked.parse(contents);
	document.getElementById('contents').innerHTML = parsed;
};
