import type { Bookmark, Editor, TinyMCE } from "tinymce";

declare global {
  interface Window {
    tinymce?: TinyMCE;
  }
}

// GOV.UK owns the controls; TinyMCE provides the editing engine and iframe.
function requireEditorElement<T extends HTMLElement = HTMLElement>(
  selector: string,
  root: ParentNode = document,
): T {
  const element = root.querySelector<T>(selector);
  if (!element) {
    throw new Error(`Missing editor element: ${selector}`);
  }
  return element;
}

function readEditorElements(
  plain: HTMLTextAreaElement,
  rich: HTMLTextAreaElement,
  wrapper: HTMLElement,
  form: HTMLFormElement,
) {
  const toolbar = requireEditorElement('[role="toolbar"]', wrapper);
  return {
    plain,
    rich,
    wrapper,
    form,
    toolbar,
    buttons: [...toolbar.querySelectorAll("button")],
    controls: [...toolbar.querySelectorAll<HTMLElement>("select, button")],
    format: requireEditorElement<HTMLSelectElement>("#editor-format", toolbar),
    linkButton: requireEditorElement<HTMLButtonElement>(
      '[data-editor-command="link"]',
      toolbar,
    ),
    panel: requireEditorElement("#editor-link-panel", wrapper),
    url: requireEditorElement<HTMLInputElement>("#editor-url", wrapper),
    urlGroup: requireEditorElement("#editor-url-group", wrapper),
    urlError: requireEditorElement("#editor-url-error", wrapper),
    linkText: requireEditorElement<HTMLInputElement>(
      "#editor-link-text",
      wrapper,
    ),
    applyLink: requireEditorElement<HTMLButtonElement>(
      "#editor-link-apply",
      wrapper,
    ),
    cancelLink: requireEditorElement<HTMLButtonElement>(
      "#editor-link-cancel",
      wrapper,
    ),
    removeLink: requireEditorElement<HTMLButtonElement>(
      "#editor-link-remove",
      wrapper,
    ),
    fallback: requireEditorElement("#editor-fallback"),
    hint: requireEditorElement("#body-hint"),
  };
}

type EditorElements = ReturnType<typeof readEditorElements>;
type GetEditor = () => Editor | undefined;

function focusToolbar(ui: EditorElements, button: HTMLElement) {
  ui.controls.forEach((control) => {
    control.tabIndex = control === button ? 0 : -1;
  });
  button.focus();
}

function setLinkError(ui: EditorElements, invalid: boolean) {
  ui.urlError.hidden = !invalid;
  ui.url.classList.toggle("govuk-input--error", invalid);
  ui.urlGroup.classList.toggle("govuk-form-group--error", invalid);
  if (invalid) {
    ui.urlError.textContent =
      "Gib eine gültige Link-Adresse mit https://, http:// oder mailto: ein.";
    ui.url.setAttribute("aria-invalid", "true");
    ui.url.setAttribute("aria-describedby", ui.urlError.id);
  } else {
    ui.url.removeAttribute("aria-invalid");
    ui.url.removeAttribute("aria-describedby");
  }
}

function parseLinkAddress(value: string): string | undefined {
  const trimmed = value.trim();
  const address =
    trimmed && !/^[a-z][a-z\d+.-]*:/i.test(trimmed)
      ? `https://${trimmed}`
      : trimmed;
  try {
    const { protocol } = new URL(address);
    if (["https:", "http:", "mailto:"].includes(protocol)) return address;
  } catch {
    // Invalid addresses are shown as field errors, preserving the selection.
  }
  return undefined;
}

function createLinkPanel(
  ui: EditorElements,
  getEditor: GetEditor,
  engine: TinyMCE,
) {
  let selection: { editor: Editor; bookmark: Bookmark } | undefined;

  const close = () => {
    ui.panel.hidden = true;
    ui.linkButton.setAttribute("aria-expanded", "false");
    selection = undefined;
  };
  const cancel = () => {
    close();
    focusToolbar(ui, ui.linkButton);
  };
  const open = () => {
    const editor = getEditor();
    if (!editor) return;
    editor.focus();
    const anchor = editor.dom.getParent(editor.selection.getNode(), "a");
    if (anchor) editor.selection.select(anchor);
    selection = { editor, bookmark: editor.selection.getBookmark(2, true) };
    ui.url.value = anchor?.getAttribute("href") ?? "";
    ui.linkText.value = anchor
      ? anchor.textContent
      : editor.selection.getContent({ format: "text" });
    ui.removeLink.hidden = !anchor;
    setLinkError(ui, false);
    ui.panel.hidden = false;
    ui.linkButton.setAttribute("aria-expanded", "true");
    ui.url.focus();
  };
  const apply = () => {
    const editor = getEditor();
    if (!editor || selection?.editor !== editor) return;
    const address = parseLinkAddress(ui.url.value);
    setLinkError(ui, !address);
    if (!address) {
      ui.url.focus();
      return;
    }
    editor.focus();
    editor.selection.moveToBookmark(selection.bookmark);
    const text = engine.DOM.encode(ui.linkText.value.trim() || address);
    editor.insertContent(editor.dom.createHTML("a", { href: address }, text));
    close();
    editor.focus();
  };
  const remove = () => {
    const editor = getEditor();
    if (!editor || selection?.editor !== editor) return;
    editor.focus();
    editor.selection.moveToBookmark(selection.bookmark);
    editor.execCommand("unlink");
    close();
    editor.focus();
  };

  ui.applyLink.addEventListener("click", apply);
  ui.cancelLink.addEventListener("click", cancel);
  ui.removeLink.addEventListener("click", remove);
  ui.panel.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      event.preventDefault();
      cancel();
    } else if (
      event.key === "Enter" &&
      event.target instanceof HTMLInputElement
    ) {
      event.preventDefault();
      apply();
    }
  });
  return { open, close, apply };
}

type LinkPanel = ReturnType<typeof createLinkPanel>;

function toolbarTarget(ui: EditorElements, event: KeyboardEvent) {
  const index = ui.controls.findIndex(
    (control) => control === document.activeElement,
  );
  switch (event.key) {
    case "ArrowRight":
      return ui.controls[(index + 1) % ui.controls.length];
    case "ArrowLeft":
      return ui.controls[(index - 1 + ui.controls.length) % ui.controls.length];
    case "Home":
      return event.target !== ui.format ? ui.controls.at(0) : undefined;
    case "End":
      return event.target !== ui.format ? ui.controls.at(-1) : undefined;
    default:
      return undefined;
  }
}

function bindToolbar(
  ui: EditorElements,
  getEditor: GetEditor,
  links: LinkPanel,
) {
  ui.toolbar.addEventListener("keydown", (event) => {
    const target = toolbarTarget(ui, event);
    if (!target) return;
    event.preventDefault();
    focusToolbar(ui, target);
  });
  ui.format.addEventListener("change", () => {
    const editor = getEditor();
    if (!editor) return;
    links.close();
    editor.focus();
    editor.execCommand("FormatBlock", false, ui.format.value);
  });
  ui.buttons.forEach((button) => {
    button.addEventListener("click", () => {
      const editor = getEditor();
      if (!editor || button.getAttribute("aria-disabled") === "true") return;
      if (button === ui.linkButton) {
        links.open();
        return;
      }
      links.close();
      editor.focus();
      const command = button.dataset.editorCommand;
      if (command === "quote") {
        editor.formatter.toggle("blockquote");
      } else if (command) {
        editor.execCommand(command);
      }
    });
  });
}

function updateToolbar(ui: EditorElements, editor: Editor) {
  if (!editor.initialized) return;
  const block = editor.dom.getParent(
    editor.selection.getNode(),
    "h2,h3,p",
    editor.getBody(),
  );
  ui.format.value =
    block && ["H2", "H3"].includes(block.nodeName)
      ? block.nodeName.toLowerCase()
      : "p";
  ui.buttons.forEach((button) => {
    const command = button.dataset.editorCommand;
    if (command && button.hasAttribute("aria-pressed")) {
      const pressed =
        command === "quote"
          ? editor.formatter.match("blockquote")
          : editor.queryCommandState(command);
      button.setAttribute("aria-pressed", String(pressed));
    }
    if (command === "Undo")
      button.setAttribute(
        "aria-disabled",
        String(!editor.undoManager.hasUndo()),
      );
    if (command === "Redo")
      button.setAttribute(
        "aria-disabled",
        String(!editor.undoManager.hasRedo()),
      );
  });
}

function describeEditor(ui: EditorElements, editor: Editor) {
  const area = editor.getBody();
  area.setAttribute("role", "textbox");
  area.setAttribute("aria-multiline", "true");
  area.setAttribute("aria-label", "Beitragstext");
  const hint = area.ownerDocument.createElement("p");
  hint.id = "rich-hint";
  hint.hidden = true;
  hint.textContent = ui.hint.textContent;
  const error = document.querySelector("#body-error");
  if (error) {
    area.setAttribute("aria-invalid", "true");
    hint.textContent += ` Fehler: ${error.textContent}`;
    editor.getContainer().classList.add("app-editor--error");
  }
  area.ownerDocument.body.after(hint);
  area.setAttribute("aria-describedby", hint.id);
}

function bindEditorFocus(getEditor: GetEditor) {
  requireEditorElement('label[for="body"]').addEventListener(
    "click",
    (event) => {
      const editor = getEditor();
      if (!editor) return;
      event.preventDefault();
      editor.focus();
    },
  );
  document
    .querySelectorAll('.govuk-error-summary a[href="#body"]')
    .forEach((link) => {
      link.addEventListener(
        "click",
        (event) => {
          const editor = getEditor();
          if (!editor) return;
          event.preventDefault();
          event.stopImmediatePropagation();
          editor.focus();
        },
        true,
      );
    });
}

function bindEditorForm(
  ui: EditorElements,
  getEditor: GetEditor,
  links: LinkPanel,
) {
  ui.form.addEventListener("submit", (event) => {
    const editor = getEditor();
    if (!editor) return;
    if (!ui.panel.hidden) {
      event.preventDefault();
      links.apply();
      return;
    }
    editor.save();
  });
}

function startEditor(ui: EditorElements, engine: TinyMCE) {
  let active: Editor | undefined;
  const getEditor = () => active;
  const initialPlain = ui.plain.value;
  const links = createLinkPanel(ui, getEditor, engine);
  const fallback = () => {
    active = undefined;
    links.close();
    ui.rich.disabled = true;
    ui.wrapper.hidden = true;
    ui.plain.disabled = false;
    ui.plain.hidden = false;
    ui.fallback.hidden = false;
  };

  bindToolbar(ui, getEditor, links);
  bindEditorForm(ui, getEditor, links);
  bindEditorFocus(getEditor);
  ui.rich.disabled = false;
  ui.rich.name = "";
  ui.wrapper.hidden = false;

  const setup = (editor: Editor) => {
    editor.on("NodeChange ExecCommand Undo Redo AddUndo", () =>
      updateToolbar(ui, editor),
    );
    editor.on("focus", () =>
      editor.getContainer().classList.add("app-editor--focused"),
    );
    editor.on("blur", () =>
      editor.getContainer().classList.remove("app-editor--focused"),
    );
    // TinyMCE restores the textarea's attributes during remove, before detach.
    editor.on("detach", fallback);
    editor.on("init", () => {
      active = editor;
      if (ui.plain.value !== initialPlain) {
        editor.setContent(
          `<p>${engine.DOM.encode(ui.plain.value).replace(/\n/g, "<br>")}</p>`,
        );
      }
      ui.rich.name = "rich_body";
      ui.plain.disabled = true;
      ui.plain.hidden = true;
      ui.toolbar.hidden = false;
      ui.fallback.hidden = true;
      describeEditor(ui, editor);
      updateToolbar(ui, editor);
    });
    editor.addShortcut("alt+f10", "Text formatieren", () => {
      focusToolbar(
        ui,
        ui.controls.find((control) => control.tabIndex === 0) ?? ui.format,
      );
    });
    editor.addShortcut("meta+k", "Link einfügen oder bearbeiten", links.open);
  };

  void engine
    .init({
      target: ui.rich,
      base_url: "/assets/tinymce",
      suffix: ".min",
      license_key: "gpl",
      language: "de",
      language_url: "/assets/tinymce/langs8/de.js",
      plugins: "lists",
      menubar: false,
      toolbar: false,
      statusbar: false,
      contextmenu: false,
      height: 320,
      resize: false,
      highlight_on_focus: false,
      promotion: false,
      valid_elements: "p,h2,h3,br,strong/b,em/i,ul,ol,li,blockquote,a[href]",
      content_css: ["/assets/govuk.css", "/assets/editor.css"],
      iframe_aria_text: "Beitragstext",
      setup,
    })
    .catch(fallback);
}

function initEditor() {
  const plain = document.querySelector<HTMLTextAreaElement>("#body");
  const rich = document.querySelector<HTMLTextAreaElement>("#rich-body");
  const wrapper = document.querySelector<HTMLElement>("#rich-editor");
  const engine = window.tinymce;
  if (!plain || !rich || !wrapper || !plain.form || !engine) return;
  startEditor(readEditorElements(plain, rich, wrapper, plain.form), engine);
}

// Wait for the deferred TinyMCE script as well as the server-rendered markup.
if (document.readyState !== "complete") {
  document.addEventListener("DOMContentLoaded", initEditor, { once: true });
} else {
  initEditor();
}
