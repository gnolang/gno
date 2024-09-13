/**
 * Replaces @username by [@username](/r/demo/users:username)
 * @param string rawData text to render usernames in
 * @returns string rendered text
 */
function renderUsernames(raw) {
  return raw.replace(/( |\n)@([_a-z0-9]{5,16})/, "$1[@$2](/r/demo/users:$2)");
}

function parseContent(source, isCode) {
  if (isCode) {
    // replace &amp; with & in code blocks
    source = source.replace(/&amp;/g, "&");

    const highlightedCode = hljs.highlightAuto(source).value;

    // Split the highlighted code into lines
    const lines = highlightedCode.split('\n');

    // Add line numbers to each line
    const numberedLines = lines.map((line, index) => {
      return `<asp class="number">${index + 1}</asp> ${line}`;
    });

    // Join the lines back into a single string
    const numberedCode = numberedLines.join('\n');

    const parser = new DOMParser();
    const doc = parser.parseFromString(numberedCode, "text/html");

    // get all span nodes of class hljs-keyword and a value of 'import'
    const nodes = doc.querySelectorAll("span.hljs-keyword");
    for (const node of nodes) {
      if (node.textContent === "import") {
        let nextNode = node;
        while (true) {
          nextNode = nextNode.nextSibling;
          if (nextNode) {
            if (nextNode.textContent.includes(")")) {
              break;
            } else if (nextNode.textContent.includes("/p") || nextNode.textContent.includes("/r")) {
              const cleanPath = nextNode.textContent.replace(/(https?:\/\/)?gno\.land\/p\//, "/p/").replace(/^"|"$/g, '');
              const link = document.createElement("a");
              link.href = cleanPath;
              link.className = "hljs-link";
              link.appendChild(nextNode.cloneNode(true));
              nextNode.replaceWith(link);
              nextNode = link;
            }
          } else {
            break;
          }
        }
      }
    }

    const codeElement = document.createElement("code");
    codeElement.classList.add("hljs");
    codeElement.innerHTML = doc.body.innerHTML;

    const preElement = document.createElement("pre");
    preElement.appendChild(codeElement);

    return preElement;
  } else {
    const { markedHighlight } = globalThis.markedHighlight;
    const { Marked } = globalThis.marked;
    const markedInstance = new Marked(
      markedHighlight({
        langPrefix: "language-",
        highlight(code, lang, info) {
          if (lang === "json") {
            try {
              code = JSON.stringify(JSON.parse(code), null, 2);
            } catch {
                console.error('Error: The provided JSON code is invalid.');
            }
          }
          const language = hljs.getLanguage(lang) ? lang : "plaintext";
          return hljs.highlight(code, { language }).value;
        },
      })
    );
    markedInstance.setOptions({ gfm: true });
    const doc = new DOMParser().parseFromString(source, "text/html");
    const contents = doc.documentElement.textContent;

    return markedInstance.parse(contents);
  }
}

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
  const accordions = Array.from(document.querySelectorAll(".accordion-trigger"));
  for (let accordion of accordions) {
    new Accordion(accordion);
  }

  const tablists = Array.from(document.querySelectorAll("[role=tablist].tabs"));
  for (let tab of tablists) {
    new Tabs(tab);
  }
});
