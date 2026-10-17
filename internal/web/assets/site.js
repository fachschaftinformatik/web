import { initAll } from "./govuk.js";

// GOV.UK's responsive navigation, checkboxes and radios.
const frontendOptions = {
  fileUpload: {
    i18n: {
      chooseFilesButton: "Datei auswählen",
      dropInstruction: "oder hier ablegen",
      noFileChosen: "Keine Datei ausgewählt",
      multipleFilesChosen: { one: "%{count} Datei ausgewählt", other: "%{count} Dateien ausgewählt" },
      enteredDropZone: "Datei im Ablagebereich.",
      leftDropZone: "Ablagebereich verlassen.",
    },
  },
  passwordInput: {
    i18n: {
      showPassword: "Anzeigen",
      hidePassword: "Verbergen",
      showPasswordAriaLabel: "Passwort anzeigen",
      hidePasswordAriaLabel: "Passwort verbergen",
      passwordShownAnnouncement: "Dein Passwort ist sichtbar.",
      passwordHiddenAnnouncement: "Dein Passwort ist verborgen.",
    },
  },
};
initAll(frontendOptions);

// Voting remains a normal POST form without JavaScript. Enhance it in place
// so keyboard focus and the reader's position in the discussion are retained.
document.querySelectorAll("form.app-vote").forEach((form) => {
  form.addEventListener("submit", async (event) => {
    const button = event.submitter;
    if (!button) return;
    event.preventDefault();
    if (form.dataset.pending) return;
    form.dataset.pending = "true";
    const buttons = [...form.querySelectorAll("button")];
    const feedback = form.parentElement.querySelector("[data-vote-feedback]");
    feedback.className = "govuk-body-s govuk-!-margin-top-1 govuk-!-margin-bottom-0";
    feedback.textContent = "Wird gespeichert …";
    buttons.forEach((item) => item.setAttribute("aria-disabled", "true"));
    try {
      const body = new URLSearchParams(new FormData(form));
      body.set("vote", button.value);
      const response = await fetch(form.action, {
        method: "POST", body, headers: { Accept: "application/json" },
      });
      if (response.redirected) { window.location.assign(response.url); return; }
      if (!response.ok) throw new Error("vote failed");
      const result = await response.json();
      form.querySelector("[data-vote-score]").textContent = result.score;
      for (const item of buttons) {
        const up = item.classList.contains("app-vote__button--up");
        const selected = result.selected === (up ? 1 : -1);
        item.setAttribute("aria-pressed", String(selected));
        item.value = selected ? "0" : (up ? "1" : "-1");
      }
      feedback.className = "govuk-visually-hidden";
      const state = result.selected === 0 ? "Bewertung zurückgenommen" : (result.selected === 1 ? "Positiv bewertet" : "Negativ bewertet");
      feedback.textContent = `${form.dataset.voteLabel}: ${state}. Punktestand: ${result.score}.`;
    } catch {
      feedback.className = "govuk-error-message govuk-!-margin-top-1 govuk-!-margin-bottom-0";
      feedback.textContent = "Die Bewertung konnte nicht gespeichert werden. Bitte versuche es erneut.";
    } finally {
      delete form.dataset.pending;
      buttons.forEach((item) => item.removeAttribute("aria-disabled"));
    }
  });
});

// Keep the upload's accessible name in the same order as its visible text.
// GOV.UK 6.6.0 inserts a hidden comma into the middle of the button contents
// but references it before those contents, breaking label-in-name checks.
function fixUploadLabels(root) {
root.querySelectorAll(".govuk-file-upload-button").forEach((button) => {
  const separatorID = `${button.id}-comma`;
  button.setAttribute(
    "aria-labelledby",
    button.getAttribute("aria-labelledby").split(" ").filter((id) => id !== separatorID).join(" "),
  );
  document.getElementById(separatorID)?.remove();
});
}
fixUploadLabels(document);

function initAssignmentSearch(root) {
  const search = root.querySelector("[data-assignment-search]");
  if (!search) return;
  const input = search.querySelector("input");
  const modules = [...root.querySelectorAll("[data-assignment-module]")];
  const count = search.querySelector("[data-assignment-count]");
  const list = root.querySelector("#assignment-modules");
  const pageSize = 10;
  let page = 0;
  const pagination = document.createElement("nav");
  pagination.className = "govuk-pagination app-module-pagination";
  pagination.setAttribute("aria-label", "Modulseiten");
  list.after(pagination);
  const empty = document.createElement("p");
  empty.className = "govuk-body";
  empty.textContent = "Keine passenden Module gefunden.";
  empty.hidden = true;
  list.after(empty);
  search.hidden = false;
  const render = () => {
    const query = input.value.trim().toLocaleLowerCase("de");
    const matches = modules.filter((module) => module.querySelector("label").textContent.toLocaleLowerCase("de").includes(query));
    const pages = Math.max(1, Math.ceil(matches.length / pageSize));
    page = Math.min(page, pages - 1);
    const visible = new Set(matches.slice(page * pageSize, (page + 1) * pageSize));
    modules.forEach((module) => { module.hidden = !visible.has(module); });
    empty.hidden = matches.length > 0;
    const selected = modules.filter((module) => module.querySelector("input:checked")).length;
    count.textContent = `${matches.length} Module gefunden. ${selected} ausgewählt.${pages > 1 ? ` Seite ${page + 1} von ${pages}.` : ""}`;
    pagination.replaceChildren();
    pagination.hidden = pages <= 1;
    const button = (label, target, accessibleLabel) => {
      const element = document.createElement("button");
      element.type = "button";
      element.className = "govuk-link govuk-pagination__link";
      element.textContent = label;
      element.dataset.page = target;
      element.setAttribute("aria-label", accessibleLabel);
      element.setAttribute("aria-controls", "assignment-modules");
      return element;
    };
    const directionButton = (label, target, direction) => {
      const element = button("", target, `${label} Modulseite`);
      const title = document.createElement("span");
      title.className = "govuk-pagination__link-title";
      title.textContent = label;
      const icon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
      icon.setAttribute("class", `govuk-pagination__icon govuk-pagination__icon--${direction}`);
      icon.setAttribute("viewBox", "0 0 15 13");
      icon.setAttribute("aria-hidden", "true");
      icon.setAttribute("focusable", "false");
      const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
      path.setAttribute("d", direction === "prev"
        ? "m6.5938-0.0078125-6.7266 6.7266 6.7441 6.4062 1.377-1.449-4.1856-3.9768h12.896v-2h-12.984l4.2931-4.293-1.414-1.414z"
        : "m8.107-0.0078125-1.4136 1.414 4.2926 4.293h-12.986v2h12.896l-4.1855 3.9766 1.377 1.4492 6.7441-6.4062-6.7246-6.7266z");
      icon.append(path);
      element.append(...(direction === "prev" ? [icon, title] : [title, icon]));
      return element;
    };
    if (page > 0) {
      const previous = document.createElement("div");
      previous.className = "govuk-pagination__prev";
      previous.append(directionButton("Vorherige", page - 1, "prev"));
      pagination.append(previous);
    }
    const numbers = document.createElement("ul");
    numbers.className = "govuk-pagination__list";
    let previousPage = -1;
    for (let i = 0; i < pages; i++) {
      if (pages > 5 && i !== 0 && i !== pages - 1 && Math.abs(i - page) > 1) continue;
      if (previousPage >= 0 && i - previousPage > 1) {
        const ellipsis = document.createElement("li");
        ellipsis.className = "govuk-pagination__item govuk-pagination__item--ellipsis";
        ellipsis.textContent = "⋯";
        ellipsis.setAttribute("aria-hidden", "true");
        numbers.append(ellipsis);
      }
      previousPage = i;
      const item = document.createElement("li");
      item.className = "govuk-pagination__item";
      const link = button(String(i + 1), i, `Seite ${i + 1}`);
      if (i === page) {
        item.classList.add("govuk-pagination__item--current");
        link.setAttribute("aria-current", "page");
      }
      item.append(link);
      numbers.append(item);
    }
    pagination.append(numbers);
    if (page < pages - 1) {
      const next = document.createElement("div");
      next.className = "govuk-pagination__next";
      next.append(directionButton("Nächste", page + 1, "next"));
      pagination.append(next);
    }
  };
  input.addEventListener("input", () => { page = 0; render(); });
  list.addEventListener("change", render);
  pagination.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-page]");
    if (!button) return;
    page = Number(button.dataset.page);
    render();
    root.querySelector("#module_id").focus();
  });
  // After validation, start on the page containing the first selected module.
  const selectedIndex = modules.findIndex((module) => module.querySelector("input:checked"));
  if (selectedIndex >= 0) page = Math.floor(selectedIndex / pageSize);
  render();
}
initAssignmentSearch(document);

// Keep the user's local files when server-side validation returns the form.
// Files stay in the browser until a valid draft has been saved.
const uploadResponses = new WeakMap();
document.addEventListener("htmx:beforeRequest", (event) => {
  if (!event.detail.elt.matches("form.app-archive-upload")) return;
  uploadResponses.set(event.detail.xhr, [...event.detail.elt.querySelector('[name="files"]').files]);
});
document.addEventListener("htmx:beforeSwap", (event) => {
  if (!uploadResponses.has(event.detail.xhr) || event.detail.xhr.status !== 422) return;
  event.detail.shouldSwap = true;
  event.detail.isError = false;
});
document.addEventListener("htmx:afterSwap", (event) => {
  const files = uploadResponses.get(event.detail.xhr);
  if (!files) return;
  uploadResponses.delete(event.detail.xhr);
  const root = document.getElementById("main-content");
  const input = root.querySelector('input[name="files"]');
  if (input) {
    const transfer = new DataTransfer();
    files.forEach((file) => transfer.items.add(file));
    input.files = transfer.files;
  }
  initAll({ ...frontendOptions, scope: root });
  fixUploadLabels(root);
  initAssignmentSearch(root);
  input?.dispatchEvent(new Event("change", { bubbles: true }));
});

// A normal confirmation page remains available without JavaScript or <dialog>.
document.querySelectorAll("[data-dialog]").forEach((trigger) => {
  const dialog = document.getElementById(trigger.dataset.dialog);
  if (!dialog || typeof dialog.showModal !== "function") return;
  trigger.setAttribute("role", "button");
  trigger.setAttribute("aria-haspopup", "dialog");
  trigger.addEventListener("click", (event) => {
    if (event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    dialog.showModal();
    dialog.querySelector("h2").focus();
  });
  const cancel = dialog.querySelector("[data-dialog-cancel]");
  const close = document.createElement("button");
  close.type = "button";
  close.className = "govuk-button govuk-button--secondary";
  close.textContent = cancel.textContent;
  cancel.replaceWith(close);
  close.addEventListener("click", () => dialog.close());
  dialog.addEventListener("close", () => trigger.focus());
  // Native dialogs keep the page inert; wrap the two action buttons explicitly
  // so Chromium does not move Tab focus into browser chrome at the boundary.
  dialog.addEventListener("keydown", (event) => {
    if (event.key !== "Tab") return;
    const actions = [...dialog.querySelectorAll("button:not([disabled]), a[href]")];
    const first = actions[0];
    const last = actions.at(-1);
    if (event.shiftKey && (document.activeElement === first || document.activeElement === dialog.querySelector("h2"))) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  });
});

const carousel = document.querySelector(".app-event-carousel");
if (carousel) {
  const slides = [...carousel.querySelectorAll(".app-event-slide")];
  const controls = carousel.querySelector(".app-event-carousel__controls");
  if (slides.length > 1) {
    let current = 0;
    let playing = false;
    let timer;
    let moving = false;
    const motion = window.matchMedia("(prefers-reduced-motion: reduce)");
    const viewport = carousel.querySelector("#event-slides");
    const track = document.createElement("div");
    track.className = "app-event-carousel__track app-event-carousel__track--instant";
    const rotation = controls.querySelector("[data-rotation]");
    const dots = [...controls.querySelectorAll("[data-slide]")];
    const status = controls.querySelector('[role="status"]');
    carousel.setAttribute("aria-roledescription", "Karussell");
    slides.forEach((slide, index) => {
      slide.hidden = false;
      slide.setAttribute("role", "group");
      slide.setAttribute("aria-roledescription", "Folie");
      slide.setAttribute(
        "aria-label",
        `${index + 1} von ${slides.length}: ${slide.dataset.title}`,
      );
      track.append(slide);
    });
    // Visual bookends let the track wrap in either direction without exposing
    // duplicate links to keyboards or assistive technology.
    const bookend = (slide) => {
      const clone = slide.cloneNode(true);
      clone.setAttribute("aria-hidden", "true");
      clone.inert = true;
      clone.dataset.clone = "true";
      return clone;
    };
    track.prepend(bookend(slides.at(-1)));
    track.append(bookend(slides[0]));
    viewport.append(track);
    carousel.classList.add("app-event-carousel--enhanced");
    const position = (index) => {
      track.style.transform = `translateX(-${index * 100}%)`;
    };
    const settle = () => {
      track.classList.add("app-event-carousel__track--instant");
      position(current + 1);
      moving = false;
    };
    const update = () => {
      slides.forEach((slide, index) => {
        slide.inert = index !== current;
        slide.setAttribute("aria-hidden", String(index !== current));
        dots[index].setAttribute("aria-disabled", String(index === current));
      });
      status.textContent = `${current + 1} von ${slides.length}`;
    };
    const show = (index) => {
      if (moving || index === current) return;
      const target = index + 1;
      current = (index + slides.length) % slides.length;
      update();
      if (motion.matches) {
        settle();
      } else {
        moving = true;
        // Commit the resting position before enabling the next transition.
        void track.offsetWidth;
        track.classList.remove("app-event-carousel__track--instant");
        position(target);
      }
    };
    track.addEventListener("transitionend", (event) => {
      if (event.target === track && event.propertyName === "transform") settle();
    });
    track.addEventListener("transitioncancel", settle);
    const setPlaying = (value) => {
      playing = value;
      window.clearInterval(timer);
      rotation.setAttribute("aria-label", playing ? "Diashow pausieren" : "Diashow starten");
      rotation.textContent = playing ? "Pausieren" : "Starten";
      status.setAttribute("aria-live", playing ? "off" : "polite");
      if (playing) timer = window.setInterval(() => show(current + 1), 7000);
    };
    controls.querySelectorAll("[data-direction]").forEach((button) => {
      button.addEventListener("click", () => {
        setPlaying(false);
        show(current + Number(button.dataset.direction));
      });
    });
    dots.forEach((button, index) => {
      button.addEventListener("click", () => {
        setPlaying(false);
        show(index);
      });
    });
    rotation.addEventListener("click", () => setPlaying(!playing));
    carousel.addEventListener("mouseenter", () => setPlaying(false));
    carousel.addEventListener("focusin", () => setPlaying(false));
    document.addEventListener("visibilitychange", () => {
      if (document.hidden) setPlaying(false);
    });
    motion.addEventListener("change", () => { setPlaying(false); settle(); });
    let touch;
    viewport.addEventListener("touchstart", (event) => {
      setPlaying(false);
      touch = event.touches.length === 1 ? { x: event.touches[0].clientX, y: event.touches[0].clientY } : null;
    }, { passive: true });
    viewport.addEventListener("touchend", (event) => {
      if (!touch || event.touches.length) { touch = null; return; }
      const dx = event.changedTouches[0].clientX - touch.x;
      const dy = event.changedTouches[0].clientY - touch.y;
      touch = null;
      if (Math.abs(dx) > 50 && Math.abs(dx) > Math.abs(dy) * 1.5) show(current + (dx < 0 ? 1 : -1));
    }, { passive: true });
    viewport.addEventListener("touchcancel", () => { touch = null; });
    update();
    settle();
    controls.hidden = false;
    setPlaying(!motion.matches && !document.hidden && !carousel.matches(":hover") && !carousel.contains(document.activeElement));
  }
}

// Return links retain the album page and focus the originating thumbnail.
const returnThumbnail = document.getElementById(window.location.hash.slice(1));
if (returnThumbnail?.matches(".app-gallery-link")) {
  returnThumbnail.focus({ preventScroll: true });
}

// Keep the image pages as ordinary links, including without JavaScript.
// Shortcuts are scoped to the viewer and never intercept text entry.
const galleryViewer = document.querySelector(".app-gallery-viewer");
if (galleryViewer) {
  document.addEventListener("keydown", (event) => {
    if (event.defaultPrevented || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey ||
        event.target.closest("input, textarea, select, [contenteditable], [role='textbox']")) return;
    const selector = { ArrowLeft: '[rel="prev"]', ArrowRight: '[rel="next"]', Escape: '.app-gallery-return' }[event.key];
    const link = selector && galleryViewer.querySelector(selector);
    if (link) {
      event.preventDefault();
      window.location.assign(link.href);
    }
  });
}

const archiveForm = document.querySelector("#archive-search");
const archiveFilters = document.querySelector(".app-archive-filters");
if (archiveFilters) {
  archiveFilters.open =
    window.matchMedia("(min-width: 48.0625em)").matches ||
    archiveFilters.dataset.active === "true";
}
archiveForm?.addEventListener("change", () => {
  let count = 0;
  for (const name of ["degree", "program", "version"]) {
    const selected = archiveForm.querySelectorAll(`[name="${name}"]:checked`).length;
    count += selected;
  }
  document.querySelector("#filter-label").textContent = count
    ? `Filter (${count})`
    : "Filter";
});

document.addEventListener("htmx:beforeRequest", (event) => {
  if (event.detail.target?.id !== "archive-results") return;
  document.querySelector("#search-error").hidden = true;
  event.detail.target.setAttribute("aria-busy", "true");
});

document.addEventListener("htmx:afterRequest", (event) => {
  if (event.detail.target?.id !== "archive-results") return;
  event.detail.target.removeAttribute("aria-busy");
  if (event.detail.failed) {
    const error = document.querySelector("#search-error");
    error.hidden = false;
  }
});

document.addEventListener("htmx:afterSwap", (event) => {
  if (event.detail.target?.id !== "archive-results") return;
  // An explicit search submission moves focus to the new result heading.
  document.querySelector("#results-heading")?.focus();
});
