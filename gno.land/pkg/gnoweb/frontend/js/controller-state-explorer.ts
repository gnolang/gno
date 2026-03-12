import { BaseController } from "./controller.js";
import { decodePkg, decodeObject, structFieldNames } from "@gnojs/amino";
import type { StateNode, QpkgResponse, QobjectResponse, QtypeResponse } from "@gnojs/amino";

export class StateExplorerController extends BaseController {
  private pkgPath = "";
  private typeCache = new Map<string, string[]>();

  protected connect(): void {
    this.pkgPath = this.getValue("pkg-path");
    const dataEl = this.getTarget("initial-data");
    if (dataEl?.textContent) {
      try {
        const raw: QpkgResponse = JSON.parse(dataEl.textContent);
        const nodes = decodePkg(raw);
        const tree = this.getTarget("tree");
        if (tree) {
          this.renderNodes(nodes, tree, 0);
          this.updateCount(nodes.length);
        }
      } catch (err) {
        console.error("Failed to parse initial state data:", err);
      }
    }
  }

  private updateCount(n: number): void {
    const countEl = this.getTarget("count");
    if (countEl) {
      countEl.textContent = `${n} top-level variable${n !== 1 ? "s" : ""}`;
    }
  }

  private renderNodes(nodes: StateNode[], container: HTMLElement, depth: number): void {
    const fragment = document.createDocumentFragment();
    for (const node of nodes) {
      fragment.appendChild(this.createRow(node, depth));
    }
    container.appendChild(fragment);
  }

  private createRow(node: StateNode, depth: number): HTMLElement {
    const row = document.createElement("div");
    row.className = "c-state-row";

    const line = document.createElement("div");
    line.className = "c-state-row__line";
    line.style.paddingLeft = `${depth * 1.25 + 0.25}rem`;

    // Toggle arrow
    const toggle = document.createElement("span");
    toggle.className = "c-state-toggle";
    if (node.expandable || (node.children && node.children.length > 0)) {
      toggle.textContent = "\u25B6";
      toggle.addEventListener("click", () => this.toggle(toggle, row, node, depth));
    }
    line.appendChild(toggle);

    // Name
    const nameEl = document.createElement("span");
    nameEl.className = "c-state-name";
    nameEl.textContent = node.name;
    line.appendChild(nameEl);

    // Separator
    line.appendChild(this.sep(":"));

    // Type
    const typeEl = document.createElement("span");
    typeEl.className = `c-state-type c-state-kind--${node.kind}`;
    typeEl.textContent = node.type;
    line.appendChild(typeEl);

    // Length
    if (node.length !== undefined && node.length > 0) {
      const lenEl = document.createElement("span");
      lenEl.className = "c-state-meta";
      lenEl.textContent = `(len=${node.length})`;
      line.appendChild(lenEl);
    }

    // Value
    if (node.value !== undefined && node.value !== "") {
      line.appendChild(this.sep("="));
      const valEl = document.createElement("span");
      valEl.className = `c-state-val c-state-val--${node.kind}`;
      valEl.textContent = node.value;
      line.appendChild(valEl);
    }

    // ObjectID (subtle, on hover)
    if (node.objectId) {
      const oidEl = document.createElement("span");
      oidEl.className = "c-state-oid";
      oidEl.textContent = node.objectId;
      oidEl.title = "Object ID \u2014 click to copy";
      oidEl.addEventListener("click", (e) => {
        e.stopPropagation();
        navigator.clipboard.writeText(node.objectId!);
        oidEl.textContent = "copied!";
        setTimeout(() => { oidEl.textContent = node.objectId!; }, 1000);
      });
      line.appendChild(oidEl);
    }

    row.appendChild(line);

    // Children container
    const kids = document.createElement("div");
    kids.className = "c-state-kids";
    if (node.children && node.children.length > 0) {
      this.renderNodes(node.children, kids, depth + 1);
    } else {
      kids.hidden = true;
    }
    row.appendChild(kids);

    return row;
  }

  private sep(char: string): HTMLElement {
    const s = document.createElement("span");
    s.className = "c-state-sep";
    s.textContent = char;
    return s;
  }

  private async toggle(toggle: HTMLElement, row: HTMLElement, node: StateNode, depth: number): Promise<void> {
    const kids = row.querySelector(".c-state-kids") as HTMLElement;
    if (!kids) return;

    const isHidden = kids.hidden;

    if (isHidden) {
      // Expand — lazy-fetch if needed
      if (kids.children.length === 0 && node.objectId) {
        toggle.classList.add("c-state-toggle--loading");
        try {
          const url = `${this.pkgPath}$state&oid=${encodeURIComponent(node.objectId)}&json`;
          const resp = await fetch(url);
          if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
          const raw: QobjectResponse = await resp.json();
          let childNodes = decodeObject(raw);

          // Resolve struct field names via qtype_json if parent has a typeId
          if (node.typeId && childNodes.length > 0) {
            childNodes = await this.resolveFieldNames(node.typeId, childNodes);
          }

          this.renderNodes(childNodes, kids, depth + 1);
        } catch (err) {
          console.error("State fetch error:", err);
          const errEl = document.createElement("span");
          errEl.className = "c-state-err";
          errEl.textContent = "Failed to load";
          kids.appendChild(errEl);
        }
        toggle.classList.remove("c-state-toggle--loading");
      }
      kids.hidden = false;
      toggle.textContent = "\u25BC";
    } else {
      kids.hidden = true;
      toggle.textContent = "\u25B6";
    }
  }

  // Fetch type info and apply struct field names to children.
  private async resolveFieldNames(typeId: string, children: StateNode[]): Promise<StateNode[]> {
    let names = this.typeCache.get(typeId);
    if (!names) {
      try {
        const url = `${this.pkgPath}$state&tid=${encodeURIComponent(typeId)}&json`;
        const resp = await fetch(url);
        if (resp.ok) {
          const raw: QtypeResponse = await resp.json();
          const resolved = structFieldNames(raw.type);
          if (resolved) {
            names = resolved;
            this.typeCache.set(typeId, names);
          }
        }
      } catch {
        // Type resolution failed — keep index-based names
      }
    }
    if (names) {
      for (let i = 0; i < children.length && i < names.length; i++) {
        children[i].name = names[i];
      }
    }
    return children;
  }
}
