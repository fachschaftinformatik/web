/* Native GOV.UK controls; TinyMCE supplies only the editing engine. */
(() => {
  const plain = document.querySelector("#body");
  const rich = document.querySelector("#rich-body");
  const wrapper = document.querySelector("#rich-editor");
  if (!plain || !rich || !window.tinymce) return;
  const form = plain.form;
  const initialPlain = plain.value;
  const toolbar = wrapper.querySelector('[role="toolbar"]');
  const buttons = [...toolbar.querySelectorAll("button")];
  const controls = [...toolbar.querySelectorAll("select, button")];
  const format = document.querySelector("#editor-format");
  const panel = document.querySelector("#editor-link-panel");
  const linkButton = toolbar.querySelector('[data-editor-command="link"]');
  const url = document.querySelector("#editor-url");
  const linkText = document.querySelector("#editor-link-text");
  const urlError = document.querySelector("#editor-url-error");
  let active, bookmark;

  const fallback = () => {
    active = undefined;
    rich.disabled = true;
    wrapper.hidden = true;
    plain.disabled = false;
    plain.hidden = false;
    document.querySelector("#editor-fallback").hidden = false;
  };
  const focusButton = (button) => {
    controls.forEach((item) => {
      item.tabIndex = item === button ? 0 : -1;
    });
    button.focus();
  };
  const closeLink = () => {
    panel.hidden = true;
    linkButton.setAttribute("aria-expanded", "false");
  };
  const clearLinkError = () => {
    urlError.hidden = true;
    url.removeAttribute("aria-invalid");
    url.removeAttribute("aria-describedby");
    url.classList.remove("govuk-input--error");
    document
      .querySelector("#editor-url-group")
      .classList.remove("govuk-form-group--error");
  };
  const openLink = () => {
    if (!active) return;
    active.focus();
    const anchor = active.dom.getParent(active.selection.getNode(), "a");
    if (anchor) active.selection.select(anchor);
    bookmark = active.selection.getBookmark(2, true);
    url.value = anchor ? anchor.getAttribute("href") : "";
    linkText.value = anchor
      ? anchor.textContent
      : active.selection.getContent({ format: "text" });
    document.querySelector("#editor-link-remove").hidden = !anchor;
    clearLinkError();
    panel.hidden = false;
    linkButton.setAttribute("aria-expanded", "true");
    url.focus();
  };
  const applyLink = () => {
    let address = url.value.trim();
    if (!/^[a-z][a-z\d+.-]*:/i.test(address) && address)
      address = "https://" + address;
    let valid = false;
    try {
      valid = ["https:", "http:", "mailto:"].includes(
        new URL(address).protocol,
      );
    } catch {
      /* Show field error below. */
    }
    if (!valid) {
      urlError.textContent =
        "Gib eine gültige Link-Adresse mit https://, http:// oder mailto: ein.";
      urlError.hidden = false;
      url.setAttribute("aria-invalid", "true");
      url.setAttribute("aria-describedby", "editor-url-error");
      url.classList.add("govuk-input--error");
      document
        .querySelector("#editor-url-group")
        .classList.add("govuk-form-group--error");
      url.focus();
      return;
    }
    active.focus();
    active.selection.moveToBookmark(bookmark);
    active.insertContent(
      active.dom.createHTML(
        "a",
        { href: address },
        tinymce.DOM.encode(linkText.value.trim() || address),
      ),
    );
    closeLink();
    active.focus();
  };

  toolbar.addEventListener("keydown", (event) => {
    const index = controls.indexOf(document.activeElement);
    let next;
    if (event.key === "ArrowRight") next = (index + 1) % controls.length;
    if (event.key === "ArrowLeft")
      next = (index - 1 + controls.length) % controls.length;
    if (event.key === "Home" && event.target !== format) next = 0;
    if (event.key === "End" && event.target !== format)
      next = controls.length - 1;
    if (next !== undefined) {
      event.preventDefault();
      focusButton(controls[next]);
    }
  });
  format.addEventListener("change", () => {
    if (!active) return;
    const block = format.value;
    closeLink();
    active.focus();
    active.execCommand("FormatBlock", false, block);
  });
  buttons.forEach((button) =>
    button.addEventListener("click", () => {
      if (!active || button.getAttribute("aria-disabled") === "true") return;
      if (button === linkButton) {
        openLink();
        return;
      }
      closeLink();
      active.focus();
      const command = button.dataset.editorCommand;
      if (command === "quote") active.formatter.toggle("blockquote");
      else active.execCommand(command);
    }),
  );
  document
    .querySelector("#editor-link-apply")
    .addEventListener("click", applyLink);
  document
    .querySelector("#editor-link-cancel")
    .addEventListener("click", () => {
      closeLink();
      focusButton(linkButton);
    });
  document
    .querySelector("#editor-link-remove")
    .addEventListener("click", () => {
      active.focus();
      active.selection.moveToBookmark(bookmark);
      active.execCommand("unlink");
      closeLink();
      active.focus();
    });
  panel.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      event.preventDefault();
      closeLink();
      focusButton(linkButton);
    }
    if (event.key === "Enter" && event.target.tagName === "INPUT") {
      event.preventDefault();
      applyLink();
    }
  });

  rich.disabled = false;
  rich.name = "";
  wrapper.hidden = false;
  tinymce
    .init({
      target: rich,
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
      setup(editor) {
        const update = () => {
          if (!editor.initialized) return;
          const block = editor.dom.getParent(
            editor.selection.getNode(),
            "h2,h3,p",
            editor.getBody(),
          );
          format.value =
            block && ["H2", "H3"].includes(block.nodeName)
              ? block.nodeName.toLowerCase()
              : "p";
          buttons.forEach((button) => {
            const command = button.dataset.editorCommand;
            if (button.hasAttribute("aria-pressed"))
              button.setAttribute(
                "aria-pressed",
                String(
                  command === "quote"
                    ? editor.formatter.match("blockquote")
                    : editor.queryCommandState(command),
                ),
              );
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
        };
        editor.on("NodeChange ExecCommand Undo Redo AddUndo", update);
        editor.on("focus", () =>
          editor.getContainer().classList.add("app-editor--focused"),
        );
        editor.on("blur", () =>
          editor.getContainer().classList.remove("app-editor--focused"),
        );
        editor.on("init", () => {
          active = editor;
          if (plain.value !== initialPlain)
            editor.setContent(
              "<p>" +
                tinymce.DOM.encode(plain.value).replace(/\n/g, "<br>") +
                "</p>",
            );
          rich.name = "rich_body";
          plain.disabled = true;
          plain.hidden = true;
          toolbar.hidden = false;
          document.querySelector("#editor-fallback").hidden = true;
          const area = editor.getBody();
          area.setAttribute("role", "textbox");
          area.setAttribute("aria-multiline", "true");
          area.setAttribute("aria-label", "Beitragstext");
          const hint = document.createElement("p");
          hint.id = "rich-hint";
          hint.hidden = true;
          hint.textContent = document.querySelector("#body-hint").textContent;
          area.ownerDocument.body.after(hint);
          const error = document.querySelector("#body-error");
          if (error) {
            area.setAttribute("aria-invalid", "true");
            hint.textContent += " Fehler: " + error.textContent;
            editor.getContainer().classList.add("app-editor--error");
          }
          area.setAttribute("aria-describedby", hint.id);
          editor.addShortcut("alt+f10", "Text formatieren", () =>
            focusButton(
              controls.find((control) => control.tabIndex === 0) || format,
            ),
          );
          editor.addShortcut(
            "meta+k",
            "Link einfügen oder bearbeiten",
            openLink,
          );
          update();
        });
        editor.on("remove", fallback);
      },
    })
    .catch(fallback);
  form.addEventListener("submit", (event) => {
    if (active) {
      if (!panel.hidden) {
        event.preventDefault();
        applyLink();
        return;
      }
      active.save();
    }
  });
  document
    .querySelector('label[for="body"]')
    .addEventListener("click", (event) => {
      if (active) {
        event.preventDefault();
        active.focus();
      }
    });
  document
    .querySelectorAll('.govuk-error-summary a[href="#body"]')
    .forEach((link) => {
      link.addEventListener(
        "click",
        (event) => {
          if (active) {
            event.preventDefault();
            event.stopImmediatePropagation();
            active.focus();
          }
        },
        true,
      );
    });
})();
