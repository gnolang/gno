// Pure helpers of the playground's files: no DOM and no editor, so node's
// test runner loads them as they are (playground-files.test.mts).
import type { EditorFile } from "@gnoweb/js/code-editor.js";

export const GNOMOD_FILE = "gnomod.toml";

// The longest name a ustar header holds without its prefix field.
const MAX_FILE_NAME_BYTES = 100;

// isValidFileName accepts a Gno source or gnomod.toml, named from a strict
// allowlist: no path part, so a name can never escape the archive Download
// builds, nothing a shell reads, and no leading dot or dash. The allowlist
// is ASCII, so the length in characters is the length in bytes.
export function isValidFileName(name: string): boolean {
	if (name.length > MAX_FILE_NAME_BYTES) return false;
	return (
		name === GNOMOD_FILE || /^[A-Za-z0-9_][A-Za-z0-9_.-]*\.gno$/.test(name)
	);
}

// shareText writes files in the share format: each under its
// "// --- name ---" header, a lone file too, so it comes back under its own
// name.
export function shareText(files: EditorFile[]): string {
	return files.map((f) => `// --- ${f.name} ---\n${f.content}`).join("\n\n");
}

// parseInitialCode reads the share format back. A header whose name is not
// a file name (or names a file already read) is a comment that only looks
// like one: it stays as text in the file it is in. Text before the first
// header, and code without headers, is main.gno.
export function parseInitialCode(code: string): EditorFile[] {
	const parts = code.split(/^\/\/ --- (.+?) ---$/m);
	if (parts.length === 1) return [{ name: "main.gno", content: code }];

	const files: EditorFile[] = [];
	if (parts[0].trim()) files.push({ name: "main.gno", content: parts[0] });
	for (let i = 1; i < parts.length; i += 2) {
		const name = parts[i].trim();
		const body = parts[i + 1];
		if (isValidFileName(name) && !files.some((f) => f.name === name)) {
			files.push({ name, content: body });
			continue;
		}
		const text = `// --- ${parts[i]} ---${body}`;
		const last = files[files.length - 1];
		if (last) last.content += text;
		else files.push({ name: "main.gno", content: parts[0] + text });
	}
	// Drop the blank lines around each header, keep one final newline.
	return files.map(({ name, content }) => ({
		name,
		content: `${content.replace(/^\n+|\n+$/g, "")}\n`,
	}));
}

// canPreview reports whether Run can show what a copy renders on the
// chain: it is a copy of a realm (from is its path, as r/demo/foo) and one
// of its files declares Render. Only an unmodified copy may, which the
// controller tracks: edited code exists nowhere the node can run it.
export function canPreview(from: string, files: EditorFile[]): boolean {
	return (
		from.startsWith("r/") &&
		files.some((f) => /^func Render\(/m.test(f.content))
	);
}

// renderText is the text a Render("") eval returned, as qeval prints a
// string, ("text" string), or the answer as it is when it is not one.
export function renderText(result: string): string {
	const m = /^\(("(?:[^"\\]|\\.)*") string\)$/s.exec(result);
	if (!m) return result;
	try {
		return JSON.parse(m[1]);
	} catch {
		return result;
	}
}

// copyFiles reads the files of a package copy, which the page hands over as
// JSON, one entry per file, so no line of a file can split it. skipped
// names the files left out: the chain takes some names isValidFileName
// does not (a leading dash, more than 100 bytes), and the reader must know
// the copy lacks them.
export function copyFiles(json: string): {
	files: EditorFile[];
	skipped: string[];
} {
	const files: EditorFile[] = [];
	const skipped: string[] = [];
	let data: unknown;
	try {
		data = JSON.parse(json);
	} catch {
		return { files, skipped };
	}
	if (!Array.isArray(data)) return { files, skipped };
	for (const f of data as Partial<EditorFile>[]) {
		const { name, content } = f ?? {};
		if (typeof name !== "string" || typeof content !== "string") continue;
		if (isValidFileName(name) && !files.some((g) => g.name === name)) {
			files.push({ name, content });
		} else {
			skipped.push(name);
		}
	}
	return { files, skipped };
}

// createTar builds a ustar archive of files, skipping any name
// isValidFileName rejects.
export function createTar(files: EditorFile[]): Uint8Array<ArrayBuffer> {
	const encoder = new TextEncoder();
	const blocks: Uint8Array[] = [];

	for (const file of files) {
		if (!isValidFileName(file.name)) {
			console.error(`skipped file with unsafe name in tar: "${file.name}"`);
			continue;
		}

		const data = encoder.encode(file.content);

		// TAR's numeric fields are ASCII octal strings
		const timestamp = Math.floor(Date.now() / 1000).toString(8);

		// Unix Standard TAR header: 512-byte fixed-field block preceding each file's data.
		// Spec: https://www.gnu.org/software/tar/manual/html_node/Standard.html
		// Fields: https://en.wikipedia.org/wiki/Tar_(computing)#UStar_format
		const header = new Uint8Array(512);

		// isValidFileName keeps the name within the field's 100 bytes.
		header.set(encoder.encode(file.name), 0);
		header.set(encoder.encode("0000644\0"), 100); // file mode
		header.set(encoder.encode("0000000\0"), 108); // uid
		header.set(encoder.encode("0000000\0"), 116); // gid
		header.set(
			encoder.encode(`${data.length.toString(8).padStart(11, "0")}\0`),
			124,
		); // file size
		header.set(encoder.encode(`${timestamp.padStart(11, "0")}\0`), 136); // mtime
		header.fill(0x20, 148, 156); // checksum placeholder (spaces)
		header[156] = 0x30; // typeflag '0' = regular file
		header.set(encoder.encode("ustar\0"), 257); // magic
		header.set(encoder.encode("00"), 263); // version

		// Sum all 512 header bytes (the 8 checksum bytes count as spaces, already
		// set above), then write the result as 6 octal digits + NUL + space
		// back into the checksum field at offset 148.
		let sum = 0;
		for (let i = 0; i < 512; i++) sum += header[i];
		header.set(encoder.encode(`${sum.toString(8).padStart(6, "0")}\0 `), 148);

		// Round up to the next 512-byte boundary because TAR requires
		// file data to be padded to a multiple of 512 bytes.
		// Extra space is padded with zeros.
		const padded = new Uint8Array(Math.ceil(data.length / 512) * 512);
		padded.set(data);

		blocks.push(header);
		blocks.push(padded);
	}

	// EOF, marked by two zero blocks
	blocks.push(new Uint8Array(1024));

	// Concatenate all blocks into a single flat byte array. Each block is
	// copied into its position in the output buffer using a running offset.
	const total = blocks.reduce((n, b) => n + b.length, 0);
	const out = new Uint8Array(total);
	let off = 0;
	for (const b of blocks) {
		out.set(b, off);
		off += b.length;
	}
	return out;
}
