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
  { name: "jumbotron", controller: "ui", toRender: (content) => `<div class="gno-jumbotron" data-gno="ui">${content}</div>` },
  { name: "stack", controller: "ui", toRender: (content) => `<div class="gno-stack" data-gno="ui">${content}</div>` },
  { name: "columns", controller: "ui", toRender: (content, attrs) => `<div class="gno-columns gno-columns-${attrs[0]}" data-gno="ui">${content}</div>` },
  { name: "box", controller: "ui", toRender: (content) => `<div data-gno="ui" class="gno-box">${content}</div>` },
  { name: "alert", controller: "ui", toRender: (content, attrs) => `<div class="gno-alert gno-alert-${attrs[0]}" role="alert" data-gno="element">${content}</div>` },
  { name: "form", controller: "ui", toRender: (content, attrs) => `<form action="${attrs[0]}" method="${attrs[1] ?? "get"}" data-gno="element">${content}</form>` },
  { name: "form-button", controller: "input", isPlain: true, toRender: (content, attrs) => `<input data-gno="input" class="gno-btn" type="${attrs[0]}" value="${content ?? ""}">` },
  {
    name: "form-input",
    controller: "input",
    isPlain: true,
    toRender: (content, attrs) =>
      `<div class="gno-input">${attrs[1] ? "<label>" + attrs[1] + "</label>" : ""}<input data-gno="input" type="${attrs[0]}" placeholder="${content ?? ""}" autocomplete="on"></div>`,
  },
  {
    name: "form-textarea",
    controller: "input",
    isPlain: true,
    toRender: (content, attrs) => `<div class="gno-input">${attrs[0] ? "<label>" + attrs[0] + "</label>" : ""}<textarea data-gno="input">${content}</textarea></div>`,
  },
  {
    name: "form-check",
    controller: "selector",
    isPlain: true,
    toRender: (content, attrs) => {
      const idfyer = (txt) => txt.replace(/\s+/g, "-");
      const els = content
        .map((item) => `<div><input type="${attrs[0]}" value="${item.text}" id="${idfyer(item.text)}"><label for="${idfyer(item.text)}">${item.text}</label></div>`)
        .reduce((a, b) => a + b, "");
      return `<div data-gno="selector" class="gno-checkboxes gno-input">${attrs[1] ? "<label>" + attrs[1] + "</label>" : ""} ${els}</div>`;
    },
  },
  {
    name: "form-select",
    controller: "selector",
    isPlain: true,
    toRender: (content, attrs) => {
      const els = content.map((item) => `<option value="${item.text}">${item.text}</option>`).reduce((a, b) => a + b, "");
      return `<div class="gno-input">${attrs[0] ? "<label>" + attrs[0] + "</label>" : ""}<select data-gno="selector" class="gno-select"> ${els}</select></div>`;
    },
  },
  {
    name: "pagination",
    controller: "ui",
    toRender: (content) => `<nav aria-label="Navigation" class="gno-pagination" data-gno="ui">${content}</nav>`,
  },
  {
    name: "breadcrumb",
    controller: "breadcrumb",
    toRender: (content) => `<nav aria-label="breadcrumb" data-gno="breadcrumb" class="gno-breadcrumb">${content}</nav>`,
  },
  {
    name: "accordion",
    controller: "accordion",
    toRender: (content, attrs) =>
      `<button type="button" aria-expanded="true" data-gno="accordion" class="gno-btn gno-accordion-trigger">${attrs[0]}</button><div role="region" class="gno-accordion-panel">${content}</div>`,
  },
  {
    name: "tabs",
    controller: "tabs",
    toRender: (content, attrs) => {
      const tabsButtons = attrs
        .map((item, i) => `<li><button role="tab" aria-selected="${i === 0 ? "true" : "false"}" aria-controls="${i}" id="${i}">${item}</button></li>`)
        .reduce((a, b) => a + b, "");
      return `<div data-gno="tabs" role="tablist" aria-labelledby="tablist-1" class="gno-tabs"><nav><ul>${tabsButtons}</ul></nav><div class="gno-jumbotron js-panel">${content}</div></div>`;
    },
  },
  {
    name: "dropdown",
    controller: "dropdown",
    toRender: (content, attrs) => `
    <div data-gno="dropdown" class="gno-dropdown">
        <button type="button" class="gno-btn" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
            ${attrs[0]}
        </button>
        ${content}
    </div>`,
  },
  {
    name: "button",
    controller: "ui",
    isPlain: true,
    toRender: (content, attrs) => (attrs[0] ? `<a class="gno-btn" role="button" href="${attrs[0]}" data-gno="ui">${content}</a>` : `<button class="gno-btn" data-gno="ui">${content}</button>`),
  },
];

/**
 *   COMPONENTS CONTROLLERS
 *   Controller classes list
 */

class GnoUi {
  static instanceCounter = 0;

  constructor(name, el) {
    this.DOM = { el };
    this.compName = name;
    this.counter = GnoUi.instanceCounter++;
  }

  mount() {
    this._setId();
  }

  _setId() {
    this.DOM.el.id = `gno-${this.compName}-${this.counter}`;
  }
}

class GnoAccordion extends GnoUi {
  constructor(name, el) {
    super(name, el);
    this._setDom();
    this._setEvents();
  }

  _setDom() {
    this.DOM.contentEl = this.DOM.el.nextElementSibling ?? this.DOM.el.parentElement.nextElementSibling;
  }

  _setEvents() {
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
      this.DOM.contentEl.classList.remove("is-hidden");
    } else {
      this.DOM.contentEl.classList.add("is-hidden");
    }
  }
}

class GnoBreadcrumb extends GnoUi {
  constructor(name, el) {
    super(name, el);
    this._setDom();
    this._setAttrs();
  }

  _setDom() {
    this.DOM.currentLink = this.DOM.el.querySelector("ul, ol").lastElementChild;
  }

  _setAttrs() {
    this.DOM.currentLink.setAttribute("aria-current", "page");
  }
}

class GnoFormElement extends GnoUi {
  constructor(name, el, i) {
    super(name, el);
    this.innerCounter = i;
  }

  mount() {
    this._setId();
    this._setNameAttr();
  }

  _setNameAttr() {
    this.DOM.el.setAttribute("name", `gno-form-${this.compName}-${this.innerCounter}`);
  }
}

class GnoInput extends GnoFormElement {
  constructor(name, el, i) {
    super(name, el, i);
  }
}

class GnoSelector extends GnoFormElement {
  constructor(name, el, i) {
    super(name, el, i);
    this._setDom();
  }
  _setDom() {
    this.DOM.checkboxes = [...this.DOM.el.querySelectorAll("input")];
  }
  _setNameAttr() {
    this.DOM.checkboxes.forEach((checkbox) => checkbox.setAttribute("name", `gno-form-${this.compName}-${this.innerCounter}`));
  }
}

class GnoDropdown extends GnoUi {
  constructor(name, el) {
    super(name, el);
    this._setDom();
    this._setEvent();
  }

  _setDom() {
    this.DOM.dropdownBtn = this.DOM.el.querySelector("button");
    this.DOM.dropdownList = this.DOM.el.querySelector("ul, ol");
  }

  _setId() {
    this.DOM.dropdownBtn.id = `gno-${this.compName}-${this.counter}`;
    this.DOM.dropdownList.setAttribute("aria-labelledby", `gno-${this.compName}-${this.counter}`);
  }

  _setEvent() {
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

class GnoTabs extends GnoUi {
  constructor(name, el) {
    super(name, el);
    this._setDom();

    this.firstTab = null;
    this.lastTab = null;
  }

  _setDom() {
    this.DOM.tabs = Array.from(this.DOM.el.querySelectorAll("[role=tab]"));
    this.DOM.tabpanels = Array.from(this.DOM.el.querySelectorAll(".js-panel > *"));
  }

  _setEvents(tab) {
    tab.addEventListener("keydown", this.onKeydown.bind(this));
    tab.addEventListener("click", this.onClick.bind(this));
  }

  mount() {
    for (let [i, tab] of this.DOM.tabs.entries()) {
      tab.tabIndex = -1;
      tab.setAttribute("aria-selected", "false");
      tab.setAttribute("aria-controls", `gno-${this.compName}-${this.counter}-${tab.getAttribute("aria-controls")}`);
      tab.id = `${this.compName}-${this.counter}-${tab.id}`;

      this.DOM.tabpanels[i].setAttribute("aria-labelledby", tab.id);

      if (!this.firstTab) {
        this.firstTab = tab;
      }
      this.lastTab = tab;
      this._setEvents(tab);
    }
    this.setSelectedTab(this.firstTab, false);
  }

  setSelectedTab(currentTab, setFocus) {
    if (typeof setFocus !== "boolean") {
      setFocus = true;
    }

    for (let i = 0; i < this.DOM.tabs.length; i += 1) {
      var tab = this.DOM.tabs[i];
      if (currentTab === tab) {
        tab.setAttribute("aria-selected", "true");
        tab.removeAttribute("tabindex");

        this.DOM.tabpanels[i].classList.remove("is-hidden");
        if (setFocus) {
          tab.focus();
        }
      } else {
        tab.setAttribute("aria-selected", "false");
        tab.tabIndex = -1;
        this.DOM.tabpanels[i].classList.add("is-hidden");
      }
    }
  }

  setSelectedToPreviousTab(currentTab) {
    let index;

    if (currentTab === this.firstTab) {
      this.setSelectedTab(this.lastTab);
    } else {
      index = this.DOM.tabs.indexOf(currentTab);
      this.setSelectedTab(this.DOM.tabs[index - 1]);
    }
  }

  setSelectedToNextTab(currentTab) {
    var index;

    if (currentTab === this.lastTab) {
      this.setSelectedTab(this.firstTab);
    } else {
      index = this.DOM.tabs.indexOf(currentTab);
      this.setSelectedTab(this.DOM.tabs[index + 1]);
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
  GnoUi,
  GnoTabs,
  GnoAccordion,
  GnoTabs,
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
      const El = new classesMap[ClassName](comp, el, i);
      El.mount(el);
    }
  }
});
