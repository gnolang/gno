/**
 * Replaces @username by [@username](/r/demo/users:username)
 * @param string rawData text to render usernames in
 * @returns string rendered text
 */
function renderUsernames(raw) {
  return raw?.replace(/( |\n)@([_a-z0-9]{5,16})/, "$1[@$2](/r/demo/users:$2)");
}

/*
 *   COMPONENTS MARKUP
 */

const components = [
  { name: "jumbotron", toRender: (content) => `<div class="jumbotron">${content}</div>` },
  { name: "stack", toRender: (content) => `<div class="stack">${content}</div>` },
  { name: "columns", toRender: (content, attrs) => `<div class="columns-${attrs[0]}">${content}</div>` },
  { name: "container", toRender: (content) => `<div>${content}</div>` },
  { name: "alert", toRender: (content, attrs) => `<div class="alert alert-${attrs[0]}" role="alert">${content}</div>` },
  { name: "form", toRender: (content, attrs) => `<form action="${attrs[0]}" method="${attrs[1] ?? "get"}">${content}</form>` },
  { name: "form-input", isPlain: true, toRender: (content, attrs) => `<input data-gno="input" type="${attrs[0]}" placeholder="${content ?? ""}">` },
  { name: "form-button", isPlain: true, toRender: (content, attrs) => `<input data-gno="input" type="${attrs[0]}" value="${content ?? ""}">` },
  { name: "form-textarea", isPlain: true, toRender: (content) => `<textarea>${content}</textarea>` },
  {
    name: "form-check",
    isPlain: true,
    toRender: (content, attrs) => {
      const idfyer = (txt) => txt.replace(/\s+/g, "-");
      const els = content
        .map((item) => `<div><input type="${attrs[0]}" value="${item.text}" id="${idfyer(item.text)}"><label for="${idfyer(item.text)}">${item.text}</label></div>`)
        .reduce((a, b) => a + b, "");
      return `<div data-gno="selectors" class="checkboxes"> ${els}</div>`;
    },
  },
  {
    name: "form-select",
    isPlain: true,
    toRender: (content) => {
      const els = content.map((item) => `<option value="${item.text}">${item.text}</option>`).reduce((a, b) => a + b, "");
      return `<select data-gno="selectors" class="select"> ${els}</select>`;
    },
  },
  {
    name: "pagination",
    toRender: (content) => `<nav aria-label="Navigation" data-gno="pagination" class="pagination">${content}</nav>`,
  },
  {
    name: "breadcrumb",
    toRender: (content) => `<nav aria-label="breadcrumb" data-gno="breadcrumb" class="breadcrumb">${content}</nav>`,
  },
  {
    name: "accordion",
    toRender: (content, attrs) =>
      `<button type="button" aria-expanded="true" data-gno="accordion" class="accordion-trigger">${attrs[0]}</button><div role="region" class="accordion-panel">${content}</div>`,
  },
  {
    name: "dropdown",
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
    isPlain: true,
    toRender: (content, attrs) => (attrs[0] ? `<a class="button" role="button" href="${attrs[0]}" data-gno="button">${content}</a>` : `<button class="button" data-gno="button">${content}</button>`),
  },
];

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
    tokenizer(src, tokens) {
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

/*
 *   COMPONENTS CONTROLERS
 */

/*
 *   ### ACCORDIONS ###
 *
 *   This content is licensed according to the W3C Software License at
 *   https://www.w3.org/Consortium/Legal/2015/copyright-software-and-document
 *
 *   Desc: Simple accordion pattern example
 */

class Accordion {
  constructor(domNode) {
    this.buttonEl = domNode;
    this.contentEl = this.buttonEl.nextElementSibling ?? this.buttonEl.parentElement.nextElementSibling;
    this.open = this.buttonEl.getAttribute("aria-expanded") === "true";

    this.buttonEl.addEventListener("click", this.onButtonClick.bind(this));
  }

  onButtonClick() {
    this.toggle(!this.open);
  }

  toggle(open) {
    if (open === this.open) {
      return;
    }

    this.open = open;

    this.buttonEl.setAttribute("aria-expanded", `${open}`);
    if (open) {
      this.contentEl.classList.remove("is-hidden");
    } else {
      this.contentEl.classList.add("is-hidden");
    }
  }
}

/*
 *   ### BREADCRUMBS ###
 *
 *   Desc: Breadcrumb ARIA attributes
 */

class Breadcrumb {
  constructor(crumbsNode) {
    this.currentLink = crumbsNode.querySelector("ul, ol").lastElementChild;
    this.currentLink.setAttribute("aria-current", "page");
  }
}

class Button {
  constructor(buttonNode, i) {
    this.DOM = {
      button: buttonNode,
    };
    this.DOM.button.id = `button-${i}`;
  }
}

/*
 *   ### FORMS ###
 *
 *   Desc: Dropdown open/close and ARIA attributes
 */

class FormInput {
  constructor(inputNode, i) {
    this.DOM = {
      input: inputNode,
    };
    this.DOM.input.setAttribute("name", `gno-form-input-${i}`);
  }
}

class FormSelector {
  constructor(checkNode, i) {
    this.DOM = {
      checkContainer: checkNode,
      checkboxes: [...checkNode.querySelectorAll("input")],
    };
    this.DOM.checkboxes.forEach((checkbox) => checkbox.setAttribute("name", `gno-form-input-${i}`));
  }
}

/*
 *   ### DROPDOWN ###
 *
 *   Desc: Dropdown open/close and ARIA attributes
 */

class Dropdown {
  constructor(dropdownNode, i) {
    this.DOM = {
      dropdown: dropdownNode,
      dropdownBtn: dropdownNode.querySelector("button"),
      dropdownList: dropdownNode.querySelector("ul, ol"),
    };

    this.isOpen = false;

    this.DOM.dropdownBtn.id = `dropdownMenuButton-${i}`;
    this.DOM.dropdownList.setAttribute("aria-labelledby", `dropdownMenuButton-${i}`);
    this.DOM.dropdownList.classList.add("is-hidden");

    document.body.addEventListener("click", (e) => {
      const DropdownBtnNode = e.target.closest("button");
      const DropdownContentNode = e.target.closest("li");

      if (!this.isOpen && !this.DOM.dropdown.contains(DropdownBtnNode)) return;
      if (!this.DOM.dropdown.contains(DropdownContentNode)) {
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

class Tabs {
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

/*
 *   ### INIT COMPONENTS ###
 */

window.addEventListener("load", function () {
  const accordions = Array.from(document.querySelectorAll('[data-gno="accordion"]'));
  for (const accordion of accordions) {
    new Accordion(accordion);
  }

  const breadcrumbs = Array.from(document.querySelectorAll('[data-gno="breadcrumb"]'));
  for (const breadcrumb of breadcrumbs) {
    new Breadcrumb(breadcrumb);
  }

  const tablists = Array.from(document.querySelectorAll("[role=tablist].tabs"));
  for (const tab of tablists) {
    new Tabs(tab);
  }

  const dropdowns = Array.from(document.querySelectorAll('[data-gno="dropdown"]'));
  for (const [i, dropdown] of dropdowns.entries()) {
    new Dropdown(dropdown, i);
  }

  const inputs = Array.from(document.querySelectorAll('[data-gno="input"]'));
  for (const [i, input] of inputs.entries()) {
    new FormInput(input, i);
  }

  const selectors = Array.from(document.querySelectorAll('[data-gno="selectors"]'));
  for (const [i, input] of selectors.entries()) {
    new FormSelector(input, i);
  }
  const buttons = Array.from(document.querySelectorAll('[data-gno="button"]'));
  for (const [i, button] of buttons.entries()) {
    new Button(button, i);
  }
});
