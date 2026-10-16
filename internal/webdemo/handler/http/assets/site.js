import { initAll } from "./govuk.js";

// GOV.UK's responsive navigation, checkboxes and radios.
initAll({
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
});

// Keep the upload's accessible name in the same order as its visible text.
// GOV.UK 6.6.0 inserts a hidden comma into the middle of the button contents
// but references it before those contents, breaking label-in-name checks.
document.querySelectorAll(".govuk-file-upload-button").forEach((button) => {
  const separatorID = `${button.id}-comma`;
  button.setAttribute(
    "aria-labelledby",
    button.getAttribute("aria-labelledby").split(" ").filter((id) => id !== separatorID).join(" "),
  );
  document.getElementById(separatorID)?.remove();
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
    carousel.setAttribute("aria-roledescription", "Karussell");
    slides.forEach((slide, index) => {
      slide.setAttribute("role", "group");
      slide.setAttribute("aria-roledescription", "Folie");
      slide.setAttribute(
        "aria-label",
        `${index + 1} von ${slides.length}: ${slide.dataset.title}`,
      );
    });
    controls.hidden = false;
    controls.querySelectorAll("button").forEach((button) => {
      button.addEventListener("click", () => {
        slides[current].hidden = true;
        current =
          (current + Number(button.dataset.direction) + slides.length) %
          slides.length;
        slides[current].hidden = false;
        carousel.querySelector('[role="status"]').textContent =
          `Album ${current + 1} von ${slides.length}`;
      });
    });
  }
}

// Return links retain the album page and focus the originating thumbnail.
const returnThumbnail = document.getElementById(window.location.hash.slice(1));
if (returnThumbnail?.matches(".app-gallery-link")) {
  returnThumbnail.focus({ preventScroll: true });
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
