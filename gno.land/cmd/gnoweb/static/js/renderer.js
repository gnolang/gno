/**
 * Replaces @username by [@username](/r/demo/users:username)
 * @param string rawData text to render usernames in
 * @returns string rendered text
 */
function renderUsernames(raw) {
  return raw?.replace(/( |\n)@([_a-z0-9]{5,16})/, "$1[@$2](/r/demo/users:$2)");
}

/**
 *   COMPONENTS MARKUP
 *   List of our components and options
 */

const components = [
  { name: "jumbotron", controller: "uiElement", toRender: (content) => `<div class="jumbotron" data-gno="uiElement">${content}</div>` },
  { name: "stack", controller: "uiElement", toRender: (content) => `<div class="stack" data-gno="uiElement">${content}</div>` },
  { name: "columns", controller: "uiElement", toRender: (content, attrs) => `<div class="columns-${attrs[0]}" data-gno="uiElement">${content}</div>` },
  { name: "container", controller: "uiElement", toRender: (content) => `<div data-gno="uiElement">${content}</div>` },
  { name: "alert", controller: "uiElement", toRender: (content, attrs) => `<div class="alert alert-${attrs[0]}" role="alert" data-gno="element">${content}</div>` },
  { name: "form", controller: "uiElement", toRender: (content, attrs) => `<form action="${attrs[0]}" method="${attrs[1] ?? "get"}" data-gno="element">${content}</form>` },
  { name: "form-input", controller: "input", isPlain: true, toRender: (content, attrs) => `<input data-gno="input" type="${attrs[0]}" placeholder="${content ?? ""}" autocomplete="on">` },
  { name: "form-button", controller: "input", isPlain: true, toRender: (content, attrs) => `<input data-gno="input" type="${attrs[0]}" value="${content ?? ""}">` },
  { name: "form-textarea", controller: "input", isPlain: true, toRender: (content) => `<textarea data-gno="input">${content}</textarea>` },
  {
    name: "form-check",
    controller: "selector",
    isPlain: true,
    toRender: (content, attrs) => {
      const idfyer = (txt) => txt.replace(/\s+/g, "-");
      const els = content
        .map((item) => `<div><input type="${attrs[0]}" value="${item.text}" id="${idfyer(item.text)}"><label for="${idfyer(item.text)}">${item.text}</label></div>`)
        .reduce((a, b) => a + b, "");
      return `<div data-gno="selector" class="checkboxes"> ${els}</div>`;
    },
  },
  {
    name: "form-select",
    controller: "selector",
    isPlain: true,
    toRender: (content) => {
      const els = content.map((item) => `<option value="${item.text}">${item.text}</option>`).reduce((a, b) => a + b, "");
      return `<select data-gno="selector" class="select"> ${els}</select>`;
    },
  },
  {
    name: "pagination",
    controller: "uiElement",
    toRender: (content) => `<nav aria-label="Navigation" class="pagination" data-gno="uiElement">${content}</nav>`,
  },
  {
    name: "breadcrumb",
    controller: "breadcrumb",
    toRender: (content) => `<nav aria-label="breadcrumb" data-gno="breadcrumb" class="breadcrumb">${content}</nav>`,
  },
  {
    name: "accordion",
    controller: "accordion",
    toRender: (content, attrs) =>
      `<button type="button" aria-expanded="true" data-gno="accordion" class="accordion-trigger">${attrs[0]}</button><div role="region" class="accordion-panel">${content}</div>`,
  },
  {
    name: "dropdown",
    controller: "dropdown",
    toRender: (content, attrs) => `
    <div data-gno="dropdown" class="dropdown">
        <button type="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
            ${attrs[0]}
        </button>
        ${content}
    </div>`,
  },
  {
    name: "button",
    controller: "uiElement",
    isPlain: true,
    toRender: (content, attrs) =>
      attrs[0] ? `<a class="button" role="button" href="${attrs[0]}" data-gno="uiElement">${content}</a>` : `<button class="button" data-gno="uiElement">${content}</button>`,
  },
];

/**
 *   COMPONENTS CONTROLLERS
 *   Controller classes list
 */

class GnoUiElement {
  constructor(el, i) {
    this.DOM = {
      el,
    };
    this.counter = i;

    this.init();
    this.setId();
  }

  init() {
    this.name = "el";
  }
  setId() {
    this.DOM.el.id = `gno-${this.name}-${this.counter}`;
  }
}

class GnoAccordion extends GnoUiElement {
  constructor(el, i) {
    super(el, i);
  }
  init() {
    this.name = "accordion";
    this.contentEl = this.DOM.el.nextElementSibling ?? this.DOM.el.parentElement.nextElementSibling;
    this.open = this.DOM.el.getAttribute("aria-expanded") === "true";

    this.DOM.el.addEventListener("click", this.onButtonClick.bind(this));
  }

  onButtonClick() {
    this.toggle(!this.open);
  }

  toggle(open) {
    if (open === this.open) {
      return;
    }

    this.open = open;

    this.DOM.el.setAttribute("aria-expanded", `${open}`);
    if (open) {
      this.DOM.el.classList.remove("is-hidden");
    } else {
      this.DOM.el.classList.add("is-hidden");
    }
  }
}

class GnoBreadcrumb extends GnoUiElement {
  constructor(el, i) {
    super(el, i);
  }
  init() {
    this.name = "breadcrumb";
    this.currentLink = this.DOM.el.querySelector("ul, ol").lastElementChild;
    this.currentLink.setAttribute("aria-current", "page");
  }
}

class GnoInput extends GnoUiElement {
  constructor(el, i) {
    super(el, i);
  }
  init() {
    this.DOM.el.setAttribute("name", `gno-form-input-${this.counter}`);
  }
}

class GnoSelector extends GnoUiElement {
  constructor(el, i) {
    super(el, i);
  }
  init() {
    this.name = "selector";
    this.DOM.checkboxes = [...this.DOM.el.querySelectorAll("input")];
    this.DOM.checkboxes.forEach((checkbox) => checkbox.setAttribute("name", `gno-form-input-${this.counter}`));
  }
}

class GnoDropdown extends GnoUiElement {
  constructor(el, i) {
    super(el, i);
  }
  setId() {
    this.DOM.dropdownBtn.id = `dropdownMenuButton-${this.counter}`;
    this.DOM.dropdownList.setAttribute("aria-labelledby", `dropdownMenuButton-${this.counter}`);
  }

  init() {
    this.name = "dropdown";
    this.DOM.dropdownBtn = this.DOM.el.querySelector("button");
    this.DOM.dropdownList = this.DOM.el.querySelector("ul, ol");

    this.isOpen = false;

    this.DOM.dropdownList.classList.add("is-hidden");

    document.body.addEventListener("click", (e) => {
      const DropdownBtnNode = e.target.closest("button");
      const DropdownContentNode = e.target.closest("li");

      if (!this.isOpen && !this.DOM.el.contains(DropdownBtnNode)) return;
      if (!this.DOM.el.contains(DropdownContentNode)) {
        this.isOpen = !this.isOpen;
      }
      this.DOM.dropdownBtn.setAttribute("aria-expanded", this.isOpen);
      this.DOM.dropdownList.classList[this.isOpen ? "remove" : "add"]("is-hidden");
    });
  }
}

/*
 *   ### TABS ###
 *
 *   This content is licensed according to the W3C Software License at
 *   https://www.w3.org/Consortium/Legal/2015/copyright-software-and-document
 *
 *   Desc: Tablist widget that implements ARIA Authoring Practices
 */

class GnoTabs {
  constructor(groupNode) {
    this.tablistNode = groupNode;

    this.tabs = [];

    this.firstTab = null;
    this.lastTab = null;

    this.tabs = Array.from(this.tablistNode.querySelectorAll("[role=tab]"));
    this.tabpanels = [];

    for (let tab of this.tabs) {
      const tabpanel = document.getElementById(tab.getAttribute("aria-controls"));

      tab.tabIndex = -1;
      tab.setAttribute("aria-selected", "false");
      this.tabpanels.push(tabpanel);

      tab.addEventListener("keydown", this.onKeydown.bind(this));
      tab.addEventListener("click", this.onClick.bind(this));

      if (!this.firstTab) {
        this.firstTab = tab;
      }
      this.lastTab = tab;
    }

    this.setSelectedTab(this.firstTab, false);
  }

  setSelectedTab(currentTab, setFocus) {
    if (typeof setFocus !== "boolean") {
      setFocus = true;
    }
    for (let i = 0; i < this.tabs.length; i += 1) {
      var tab = this.tabs[i];
      if (currentTab === tab) {
        tab.setAttribute("aria-selected", "true");
        tab.removeAttribute("tabindex");
        this.tabpanels[i].classList.remove("is-hidden");
        if (setFocus) {
          tab.focus();
        }
      } else {
        tab.setAttribute("aria-selected", "false");
        tab.tabIndex = -1;
        this.tabpanels[i].classList.add("is-hidden");
      }
    }
  }

  setSelectedToPreviousTab(currentTab) {
    let index;

    if (currentTab === this.firstTab) {
      this.setSelectedTab(this.lastTab);
    } else {
      index = this.tabs.indexOf(currentTab);
      this.setSelectedTab(this.tabs[index - 1]);
    }
  }

  setSelectedToNextTab(currentTab) {
    var index;

    if (currentTab === this.lastTab) {
      this.setSelectedTab(this.firstTab);
    } else {
      index = this.tabs.indexOf(currentTab);
      this.setSelectedTab(this.tabs[index + 1]);
    }
  }

  /* EVENT HANDLERS */

  onKeydown(event) {
    const tgt = event.currentTarget,
      flag = false;

    switch (event.key) {
      case "ArrowLeft":
        this.setSelectedToPreviousTab(tgt);
        flag = true;
        break;

      case "ArrowRight":
        this.setSelectedToNextTab(tgt);
        flag = true;
        break;

      case "Home":
        this.setSelectedTab(this.firstTab);
        flag = true;
        break;

      case "End":
        this.setSelectedTab(this.lastTab);
        flag = true;
        break;

      default:
        break;
    }

    if (flag) {
      event.stopPropagation();
      event.preventDefault();
    }
  }

  onClick(event) {
    this.setSelectedTab(event.currentTarget);
  }
}

/**
 *   COMPONENTS BUILDER
 *   Markedjs component builder
 */

const extensionBuilder = (comp) => {
  const { name, toRender, isPlain } = comp;
  const startReg = RegExp(`:::${name}`);
  const tokenizerReg = RegExp(`^:::${name}(\\s\\([^\r\n]*?\\))?(\n([\\s\\S]*?)\n?:::${name})?\/`);
  const variablesReg = /\(([^()]+)\)/g;
  return {
    name: name,
    level: "block",
    start(src) {
      return startReg.exec(src)?.index;
    },
    tokenizer(src) {
      const match = tokenizerReg.exec(src);
      if (match) {
        const token = {
          type: name,
          raw: match[0],
          text: match[3]?.trim(),
          attrs: match[1]?.match(variablesReg)?.map((attr) => attr.substring(1, attr.length - 1)) ?? [],
          tokens: [],
        };
        this.lexer.blockTokens(token.text ?? "", token.tokens);
        return token;
      }
    },
    renderer(token) {
      return toRender(isPlain ? token.tokens[0]?.items || token.text : this.parser.parse(token.tokens), token.attrs);
    },
  };
};

function parseContent(source) {
  components.forEach((comp) => marked.use({ extensions: [extensionBuilder(comp)] }));
  marked.setOptions({ gfm: true });
  const doc = new DOMParser().parseFromString(source, "text/html");
  const contents = doc.documentElement.textContent;
  return marked.parse(contents);
}

/**
 *  INIT COMPONENTS
 */

let classesMap = {
  GnoUiElement,
  GnoTabs,
  GnoAccordion,
  GnoBreadcrumb,
  GnoInput,
  GnoSelector,
  GnoDropdown,
};

window.addEventListener("load", function () {
  const controllers = [...new Set(components.map((comp) => comp.controller))];
  for (const comp of controllers) {
    if (comp === undefined) continue;
    const els = Array.from(document.querySelectorAll(`[data-gno="${comp}"]`));
    const ClassName = `Gno${comp.charAt(0).toUpperCase() + comp.slice(1)}`;

    for (const [i, el] of els.entries()) {
      new classesMap[ClassName](el, i);
    }
  }

  const tablists = Array.from(document.querySelectorAll("[role=tablist].tabs"));
  for (const tab of tablists) {
    new Tabs(tab);
  }
});
