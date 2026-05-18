import {
	defaultKeymap,
	history,
	historyKeymap,
	indentWithTab,
} from "@codemirror/commands";
import {
	bracketMatching,
	defaultHighlightStyle,
	indentOnInput,
	indentUnit,
	StreamLanguage,
	syntaxHighlighting,
} from "@codemirror/language";
import { go } from "@codemirror/legacy-modes/mode/go";
import { toml } from "@codemirror/legacy-modes/mode/toml";
import { Compartment, EditorState, type Extension } from "@codemirror/state";
import { oneDark } from "@codemirror/theme-one-dark";
import {
	drawSelection,
	EditorView,
	highlightActiveLine,
	highlightActiveLineGutter,
	keymap,
	lineNumbers,
} from "@codemirror/view";

const goLang = StreamLanguage.define(go);
const tomlLang = StreamLanguage.define(toml);

function languageFromFilename(name: string): StreamLanguage<unknown> {
	return name.endsWith(".toml") ? tomlLang : goLang;
}

function themeFor(isDarkMode: boolean): Extension {
	return isDarkMode
		? oneDark
		: syntaxHighlighting(defaultHighlightStyle, { fallback: true });
}

export function isDarkMode(): boolean {
	return document.documentElement.getAttribute("data-theme") === "dark";
}

export interface CodeEditorOptions {
	parent: HTMLElement;
	content: string;
	fileName: string;
	isDarkMode: boolean;
	onRun?: () => void;
}

export class CodeEditor {
	readonly view: EditorView;
	private readonly langCompartment = new Compartment();
	private readonly themeCompartment = new Compartment();

	constructor(opts: CodeEditorOptions) {
		const extensions: Extension[] = [
			lineNumbers(),
			highlightActiveLine(),
			highlightActiveLineGutter(),
			drawSelection(),
			history(),
			indentOnInput(),
			indentUnit.of("\t"),
			bracketMatching(),
			this.langCompartment.of(languageFromFilename(opts.fileName)),
			this.themeCompartment.of(themeFor(opts.isDarkMode)),
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

		extensions.push(
			keymap.of([indentWithTab, ...historyKeymap, ...defaultKeymap]),
		);

		this.view = new EditorView({
			parent: opts.parent,
			state: EditorState.create({
				doc: opts.content,
				extensions,
			}),
		});
	}

	getCode(): string {
		return this.view.state.doc.toString();
	}

	setCode(text: string): void {
		this.view.dispatch({
			changes: { from: 0, to: this.view.state.doc.length, insert: text },
		});
	}

	setLanguage(fileName: string): void {
		this.view.dispatch({
			effects: this.langCompartment.reconfigure(languageFromFilename(fileName)),
		});
	}

	changeTheme(isDarkMode: boolean): void {
		this.view.dispatch({
			effects: this.themeCompartment.reconfigure(themeFor(isDarkMode)),
		});
	}

	focus(): void {
		this.view.focus();
	}
}
