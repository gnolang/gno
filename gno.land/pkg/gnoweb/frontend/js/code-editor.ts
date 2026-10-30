import {
	defaultKeymap,
	history,
	historyKeymap,
	indentWithTab,
} from "@codemirror/commands";
import {
	bracketMatching,
	HighlightStyle,
	indentOnInput,
	indentUnit,
	StreamLanguage,
	syntaxHighlighting,
} from "@codemirror/language";
import { go } from "@codemirror/legacy-modes/mode/go";
import { toml } from "@codemirror/legacy-modes/mode/toml";
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search";
import { EditorState, type Extension } from "@codemirror/state";
import {
	drawSelection,
	EditorView,
	highlightActiveLine,
	highlightActiveLineGutter,
	highlightSpecialChars,
	keymap,
	lineNumbers,
} from "@codemirror/view";
import { tags } from "@lezer/highlight";

// Built-in types, which the go mode reports as plain keywords.
const goTypes = new Set(
	"bool byte complex64 complex128 float32 float64 int int8 int16 int32 int64 string uint uint8 uint16 uint32 uint64 uintptr error rune any comparable".split(
		" ",
	),
);

// The go mode, retagged to split tokens the way chroma's Go lexer does:
// types apart from keywords, and calls/declarations as function names.
const goLang = StreamLanguage.define({
	...go,
	token(stream, state) {
		const style = go.token(stream, state);
		if (style === "keyword" && goTypes.has(stream.current())) return "typeName";
		if (style === "variable" && stream.match(/^\s*\(/, false)) {
			return "variableName.function";
		}
		return style;
	},
});
const tomlLang = StreamLanguage.define(toml);

function languageFromFilename(name: string): StreamLanguage<unknown> {
	return name.endsWith(".toml") ? tomlLang : goLang;
}

// Tokens get chroma's class names, so the editor is coloured by the same
// stylesheet (/public/_chroma/style.css) as the read-only source view, in both
// themes, and never needs its own palette or a theme switch.
const chromaHighlight = HighlightStyle.define([
	{ tag: tags.keyword, class: "chroma-k" },
	{ tag: tags.typeName, class: "chroma-kt" },
	{
		tag: [
			tags.function(tags.variableName),
			tags.definition(tags.variableName),
			tags.propertyName,
		],
		class: "chroma-nf",
	},
	{ tag: [tags.atom, tags.bool, tags.null], class: "chroma-nb" },
	{ tag: tags.number, class: "chroma-m" },
	{ tag: tags.string, class: "chroma-s" },
	{ tag: tags.comment, class: "chroma-c1" },
	{ tag: tags.operator, class: "chroma-o" },
]);

// CodeMirror's parts use gnoweb's tokens; the frame around them is
// .b-code-editor-area. This lives here rather than in main.css because
// CodeMirror adds its .cm-* classes at runtime, which PurgeCSS cannot see.
const gnowebTheme = EditorView.theme({
	"&": {
		width: "100%",
		minWidth: "0",
		maxHeight: "70vh",
	},
	"&.cm-focused": { outline: "none" },
	".cm-scroller": {
		minHeight: "var(--g-space-64)",
		overflow: "auto",
		fontFamily: "var(--g-font-family-mono)",
		lineHeight: "var(--g-line-height-normal)",
	},
	".cm-content.chroma-chroma": {
		paddingBlock: "var(--g-space-3)",
		backgroundColor: "transparent",
		color: "var(--s-color-text-secondary)",
		caretColor: "var(--s-color-text-primary)",
	},
	".cm-cursor": { borderLeftColor: "var(--s-color-text-primary)" },
	".cm-gutters": {
		backgroundColor: "var(--s-color-bg-base)",
		color: "var(--s-color-text-tertiary)",
		border: "none",
		borderRight: "var(--s-divider)",
	},
	".cm-activeLine, .cm-activeLineGutter": {
		backgroundColor: "var(--s-color-bg-hover)",
	},
	// A light highlight under the text, which keeps its syntax colours: the
	// selectors match CodeMirror's own so they win, and the native selection
	// (white text on brand green elsewhere in gnoweb) stays hidden.
	"&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground, .cm-selectionLayer .cm-selectionBackground":
		{
			backgroundColor: "var(--s-color-bg-selected)",
		},
	".cm-content ::selection": {
		backgroundColor: "transparent",
		color: "inherit",
	},
	".cm-matchingBracket": {
		backgroundColor: "var(--s-color-bg-selected)",
		outline: "var(--s-divider)",
	},
	".cm-panels": {
		backgroundColor: "var(--s-color-bg-surface-secondary)",
		color: "var(--s-color-text-primary)",
	},
	".cm-panels.cm-panels-bottom": { borderTop: "var(--s-divider)" },
	".cm-textfield, .cm-button": { fontSize: "var(--g-font-size-90)" },
});

// cspNonce reads the per-response Content-Security-Policy nonce that the server
// embeds in <meta name="csp-nonce">. Under a strict CSP (style-src without
// 'unsafe-inline'), CodeMirror's runtime-injected <style> elements are blocked
// unless they carry this nonce. Returns "" when no nonce is present (e.g. the
// server runs without strict headers), in which case CodeMirror behaves as before.
function cspNonce(): string {
	return (
		document
			.querySelector<HTMLMetaElement>('meta[name="csp-nonce"]')
			?.getAttribute("content") ?? ""
	);
}

export interface EditorFile {
	name: string;
	content: string;
}

export interface CodeEditorOptions {
	parent: HTMLElement;
	content: string;
	fileName: string;
	// label names the editor for assistive technologies.
	label: string;
	onRun?: () => void;
	onChange?: () => void;
}

export class CodeEditor {
	private readonly view: EditorView;
	private readonly baseExtensions: Extension[];
	private readonly label: string;
	// One EditorState per file, in tab order, so each file keeps its own text
	// and undo history: switching tabs never lets Ctrl+Z write one file's text
	// into another. The shown file's entry is refreshed when it is left.
	private readonly states = new Map<string, EditorState>();
	fileName: string;

	constructor(opts: CodeEditorOptions) {
		const nonce = cspNonce();
		this.label = opts.label;
		this.fileName = opts.fileName;

		const extensions: Extension[] = [
			...(nonce ? [EditorView.cspNonce.of(nonce)] : []),
			lineNumbers(),
			highlightActiveLine(),
			highlightActiveLineGutter(),
			highlightSpecialChars(),
			highlightSelectionMatches(),
			drawSelection(),
			history(),
			indentOnInput(),
			indentUnit.of("\t"),
			bracketMatching(),
			syntaxHighlighting(chromaHighlight),
			gnowebTheme,
			// On the content only: chroma's .chroma-chroma rule also paints a
			// background, which the frame below must keep.
			EditorView.contentAttributes.of({ class: "chroma-chroma chroma-light" }),
		];

		if (opts.onRun) {
			const runHandler = opts.onRun;
			extensions.push(
				keymap.of([
					{
						key: "Mod-Enter",
						preventDefault: true,
						run: () => {
							runHandler();
							return true;
						},
					},
				]),
			);
		}

		if (opts.onChange) {
			const changeHandler = opts.onChange;
			extensions.push(
				EditorView.updateListener.of((update) => {
					if (update.docChanged) changeHandler();
				}),
			);
		}

		extensions.push(
			keymap.of([
				indentWithTab,
				...searchKeymap,
				...historyKeymap,
				...defaultKeymap,
			]),
		);
		this.baseExtensions = extensions;

		const state = this.createState(opts.fileName, opts.content);
		this.states.set(opts.fileName, state);
		this.view = new EditorView({ parent: opts.parent, state });
	}

	private createState(fileName: string, doc: string): EditorState {
		return EditorState.create({
			doc,
			extensions: [
				...this.baseExtensions,
				languageFromFilename(fileName),
				EditorView.contentAttributes.of({
					"aria-label": `${this.label}: ${fileName}. Press Escape then Tab to leave the editor.`,
				}),
			],
		});
	}

	private stateOf(fileName: string): EditorState | undefined {
		return fileName === this.fileName
			? this.view.state
			: this.states.get(fileName);
	}

	getCode(): string {
		return this.view.state.doc.toString();
	}

	// setCode replaces the shown file's text as one undoable change.
	setCode(text: string): void {
		this.view.dispatch({
			changes: { from: 0, to: this.view.state.doc.length, insert: text },
		});
	}

	fileNames(): string[] {
		return [...this.states.keys()];
	}

	content(fileName: string): string {
		return this.stateOf(fileName)?.doc.toString() ?? "";
	}

	files(): EditorFile[] {
		return this.fileNames().map((name) => ({
			name,
			content: this.content(name),
		}));
	}

	// addFile adds a file after the others without showing it.
	addFile(fileName: string, content: string): void {
		this.states.set(fileName, this.createState(fileName, content));
	}

	// openFile shows a file added before, with its text and undo history.
	openFile(fileName: string): void {
		const next = this.states.get(fileName);
		if (!next || fileName === this.fileName) return;
		this.states.set(this.fileName, this.view.state);
		this.fileName = fileName;
		this.view.setState(next);
	}

	// closeFile forgets a file that is not shown.
	closeFile(fileName: string): void {
		if (fileName !== this.fileName) this.states.delete(fileName);
	}
}
