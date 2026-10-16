# GOV.UK pattern review

Reviewed against the [GOV.UK pattern catalogue](https://design-system.service.gov.uk/patterns/)
on 16 October 2026 for the Go team/archive preview and a possible migration of
the existing Fachschaft site. These are design decisions, not a claim of audited
accessibility. Headless Chromium visual and interaction review is recorded below;
manual assistive-technology review is still needed.

## Approach

### Administration pattern review, 17 October 2026

Implemented in the local demo. Reviewed the five references below;
also inspected the rendered GOV.UK guidance in Playwright Chromium.

- [Check answers](https://design-system.service.gov.uk/patterns/check-answers/):
  use a review page for publishing invitations and minutes, with document links,
  meeting details, contextual change links and an explicit publish button.
  Preserve entered values and uploaded files when returning to edit.
- [Complete multiple tasks](https://design-system.service.gov.uk/patterns/complete-multiple-tasks/):
  intended for longer transactions spanning multiple sessions. Prefer a simple
  administration index and contextual navigation for these independent operations.
- [MoJ Add to a list](https://design-patterns.service.justice.gov.uk/patterns/add-to-a-list/):
  marked “To be reviewed”. Use its add/review/change/remove interaction as
  supplementary inspiration for assigning people to departments and project roles.
- [File upload](https://design-system.service.gov.uk/components/file-upload/):
  use the improved, translated GOV.UK upload control for PDFs, with specific
  missing-file, wrong-type, empty-file and file-size errors. Retain uploads
  through validation and review rather than asking users to select them again.
- [Summary list](https://design-system.service.gov.uk/components/summary-list/):
  use for individual user/member details and publication review. Use tables for
  searchable user lists; summary cards only for richer repeated records.

Suggested administration sections: Benutzer, Mitglieder, Referate, Projektstellen,
and Sitzungen und Protokolle. Separate login accounts, public member profiles,
and responsibility assignments so changes to one do not implicitly change another.
Search users by name/email with an explicit secondary search button, a linked
name in a compact table and pagination when needed. Edit records on dedicated
forms, then return to the record/list with a success notification. Publish
documents under a meeting rather than creating disconnected invitation/minutes
records. Review before publication; routine reversible edits save directly.

The local demo uses admin/admin as an explicit demo-only
credential exception, a server-side session and protected administration routes,
with logout and CSRF protection for writes. Existing email-domain validation
remains for normal accounts. Demo changes/uploads live in process memory and
appear on public demo pages; restarting resets them. Domain operations live in
account/team/meetings packages, with thin HTTP handlers. The optional `-data`
argument loads only non-secret user display columns from a read-only SQLite
snapshot; no production credentials or sessions are imported.

### Password reset confirmation dialog, 17 October 2026

- [MoJ Confirm an action](https://design-patterns.service.justice.gov.uk/patterns/confirm-an-action/)
  is Experimental (August 2026) and describes confirmation/success pages, not a
  modal component. Use its target-details and explicit-action approach. GOV.UK
  Frontend has no modal component; this dialog is a local extension.
- [WAI-ARIA modal dialog pattern](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/)
  requires a named dialog, focus inside it, contained tab order, inert background,
  Escape/cancel closure and focus restored to the invoking control.
- [MDN dialog](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/dialog)
  recommends native `showModal()` for modal behaviour and an explicit closing
  control. Use a focusable heading for initial focus because the dialog contains
  several paragraphs. The dialog itself gets no tabindex or redundant custom
  background-inert implementation. A small Tab/Shift+Tab boundary handler keeps
  focus cycling between the dialog actions rather than into Chromium's browser
  chrome; native behaviour supplies modal isolation and Escape closure.

The account detail page offers Passwort zurücksetzen as a secondary action.
Its native dialog names the destination email and offers Link zum Zurücksetzen
senden and Abbrechen. This previews sending a reset link, never a plaintext
password; existing credentials would stay valid until a new password is chosen.
Normal green/secondary buttons fit this action. A standard confirmation page
shares the same form without JavaScript or native-dialog support. GET opens the
confirmation; CSRF-protected POST previews the action and redirects to an explicit
no-email/no-password-change notice. Addresses are displayed lowercase without
changing the source database.

### User-list pagination, 17 October 2026

Reviewed [GOV.UK Pagination](https://design-system.service.gov.uk/components/pagination/)
and the rendered numbered-list example. Search results use the list variant,
with directional icons, page-number links, the current page marked with
`aria-current="page"`, first/last pages and ellipses for skipped ranges. The
current page and its neighbours remain visible on small screens. Mobile items
use 40px minimum widths so seven items fit at 320px; previous and next controls
stack using GOV.UK's responsive layout. The document title announces the current
page and total, replacing the former unstructured “Seite 1 von 19 / Weiter” row.
Search is applied before pagination and survives every page link. New searches
start on page one; invalid/out-of-range pages redirect to the first results page.
No pagination is shown for a single page or an empty result set. User confirmation
tags use the `verified` flag independently of active/inactive account status;
the visible column label is Status at the user's request.

The improved GOV.UK 6.6.0 file upload's hidden comma is removed after initialisation,
along with its aria-labelledby reference: otherwise the separator is announced
in a different order from the visible text and fails Axe label-in-name checks.
The label, status and instructions remain in the accessible name; vendor assets
are unchanged.

### Archive search layout review, 17 October 2026

References reviewed:

- [GOV.UK Checkboxes](https://design-system.service.gov.uk/components/checkboxes/):
  multiple selections; smaller controls are appropriate for search filters.
- [GOV.UK Radios](https://design-system.service.gov.uk/components/radios/):
  single selection only, rather than simply a short list of options.
- [GOV.UK Table](https://design-system.service.gov.uk/components/table/):
  use for comparing rows and columns, not general page layout; reduce the amount
  of data and split large tables across pages where possible.
- [GOV.UK Details](https://design-system.service.gov.uk/components/details/):
  intended for information only some users need, not hiding the primary downloads.
  A separate, linkable module list is used rather than hiding uploads in a set
  of disclosures or adding a modal.
- [MoJ Filter a list](https://design-patterns.service.justice.gov.uk/patterns/filter-a-list/)
  and [Filter](https://design-patterns.service.justice.gov.uk/components/filter/):
  start unfiltered, submit selected filters explicitly, combine with search.
  These are supplementary MoJ patterns marked “To be reviewed”, not a core
  GOV.UK archive-search pattern. Its rendered example was inspected in Chromium.

The archive now has a full-width heading and introduction, a prominent
“Reko einreichen” action, search above a result table, and filters alongside
the results. After re-reviewing the Table, Details and MoJ filter guidance for
modules with multiple uploads, the overview uses two columns: Modul and Rekos.
Each module appears once, with a direct PDF link for its newest matching exam.
If more than one upload matches, “Alle N Rekos ansehen” opens
`/archive/{moduleID}` with the full list of matching uploads. A single upload
requires only one PDF click; older uploads take the module-list link and one
PDF click. There is no individual exam detail page. This supersedes the earlier
five- and three-column exam-row overview layouts. The overview retains its native
two-column table at mobile widths. The module page has Datum, Studiengang,
Abschluss, Prüfungsordnung and Dokumente columns, newest exam date first. Date
values use regular font weight while retaining row-header semantics. Below 641px its
rows stack with visible labels and explicit table roles, while column headers
remain available to assistive technology. Its introduction and submit action
were removed at the user's request; the page starts with the module heading,
upload count and table. PDF links match
the meetings archive: a document-name link with hidden
context, then “(PDF, size)”, without forced downloads or a different button style.
Eight fixture uploads across six modules remain examples, including three
uploads in Algorithmen und Datenstrukturen. All “Beispiel-Reko” links use
one explicitly labelled, 1 kB example PDF generated by
`python3 tools/update-archive-example.py`; these are not production exam files.
The grouping uses explicit module IDs, not just display titles. Module counts,
the latest direct download, and the module's upload list all use the same filters.
Query text and repeated degree/programme/version parameters survive the module
link and the Altklausuren breadcrumb; a link on the module page can show all of
its uploads. Unknown module IDs return 404; known modules with no matching
uploads show an empty state. Programme filters include Informatik und Design. The submit action points
to the existing website's archive. The additional submission-explanation section
was removed at the user's request.

Abschluss (Bachelor/Master), Studiengang and Prüfungsordnung use small checkboxes:
multiple selections are possible in each group. The eight sample uploads
are Bachelor records; Master currently gives the normal empty-results state.
Within each group,
selected values are ORed; across groups and search they are ANDed. No selection
means all values. There is no “Alle” hint, checkbox or radio; checkmarks alone
communicate the selection, without a status hint appearing or disappearing.
There is no extra applied-filter summary above the result table; the form
controls retain the selected values.
This replaces the earlier single-PO radios. “Filter anwenden” and “Suchen” submit
the same GET form; no filtering happens just from changing a checkbox. The
secondary apply button sits on white rather than its own grey background, so
its shape remains visible. Search also uses secondary styling, matching the
Forum search; “Reko einreichen” is the primary page action, matching “Beitrag
verfassen” in the Forum. Reset is a link to the unfiltered page.

On mobile, search precedes a native filter disclosure and results. Filters start
open on desktop, or when selections are applied; without JavaScript they remain
open. HTMX result replacement retains the form, query URL and result-heading
focus behaviour. Native GET submission remains available without JavaScript.
Each filter group has its own fieldset and legend. This interaction still needs
manual screen-reader review.

Verification: demo Go tests and vet passed. Chromium covered the search/filter
layout at 1440, 768, 390 and 320px, multi-PO queries, retained
selections, explicit keyboard submission, reset, result focus, empty results and
no-JavaScript submission. The final removal of the changing “Alle” hints was
checked again at 1440 and 390px, including reset and multi-PO submission. Axe
reported no violations on the checked search pages. Final screenshots are
`archive-redesign-1440.png` and `archive-redesign-390.png` in the session's local
temporary screenshot directory.
The final five-column/stacked table and degree filter passed Chromium checks at
1440, 1100, 1024, 768, 390 and 320px, including preserved accessible column/row
headers, combined degree/programme/PO submission, retained degree selection,
Master empty results and Axe. Desktop/mobile screenshots were visually reviewed.

The visual reference is the [Design System documentation site's Get started page](https://design-system.service.gov.uk/get-started/),
not just the default transactional-service template. The reference was inspected
in Chromium at desktop and mobile widths. Its site search and split-pane sidebar
are custom documentation-site elements, not standalone GOV.UK Frontend components.

Keep the existing site's German content, grouped team directory and module-based
archive. Use GOV.UK Frontend 6.6.0 with standard colours, controls, typography,
focus treatments and responsive navigation. The initial Arial adaptation was
later replaced with locally embedded GDS Transport at the user's request for
this local visual prototype (see the typography review below), without government
logos or Crown branding. Retaining the old branding is not a
requirement.

Patterns describe user journeys; components provide individual UI elements.
The catalogue does not prescribe a team-directory or archive-search pattern.
Archive filtering is a local composition of standard components, to be validated
with students rather than presented as an official GOV.UK pattern.

## Minimal UI decisions

- Match the documentation shell: 1100px container, blue header, pale Service
  navigation, 260px sidebar, 30px column gap and an XL page heading.
  At 1440px this places the container at x=170 and the content at x=460, matching
  the inspected reference. Stack below the desktop breakpoint.
- Use Generic header markup for the site name, with a local blue-header/search
  adaptation. Service navigation contains the two implemented pages. No Crown or
  GOV.UK logo. The source's government-service context is not a claim about this site.
- The header search is a visual prototype: editable input with a magnifier,
  accessible label and a Demo placeholder. No form, endpoint or submit handler.
  It is separate from the working archive search.
- The team sidebar links to its two sections and is omitted on small screens,
  where those sections are already directly visible. Archive search and optional
  filters occupy the sidebar on desktop, below the page heading on mobile.
- Use Summary lists for responsibility → members/contact. These are genuine
  key/value relationships, with ordinary lists inside values where several members
  share a responsibility. Preserve roles and show shared addresses once.
- Put archive inputs directly on the page, without a panel. Keep labels and the
  programme hint; remove redundant introductory and example text.
- Show one results table with module/date and programme/PO columns. No collapsed
  content, per-record demo descriptions or repeated unavailable-download text.
- Put the short preview notice in the footer with the contact link. A Phase banner
  is not needed in this documentation-style shell.
- Keep local CSS for the documentation shell/search, request indicator, table
  overflow and long email wrapping. Use the Sass page-width setting and standard
  component classes for the rest; preserve their focus states and type scale.

Sources: [Layout](https://design-system.service.gov.uk/styles/layout/),
[Service navigation](https://design-system.service.gov.uk/components/service-navigation/),
[Summary list](https://design-system.service.gov.uk/components/summary-list/),
and [Details](https://design-system.service.gov.uk/components/details/).

Summary cards are not recommended for small amounts of related information.
Details should not hide content most users need, and is not the recommended
component for multiple important content sections. These findings supersede the
initial card grid and collapsed archive modules. The section sidebar follows the
new documentation-site reference; it does not restyle Service navigation vertically.
One Details component now contains optional programme/PO filters, not results.
Search and results are always visible; active filters keep Details open.

### Component selection

| Content                       | Component                                                                                                 | Reason                                                                                                                                                                                          |
| ----------------------------- | --------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Site identity and page links  | Service navigation                                                                                        | Independent, repeatable tasks; native mobile toggle and active state.                                                                                                                           |
| Site header                   | [Generic header](https://design-system.service.gov.uk/components/generic-header/) with local shell styles | Site name and blue background; the separate search field follows the documentation-site reference.                                                                                              |
| Preview status                | Plain footer text                                                                                         | A short prototype notice, without an additional banner.                                                                                                                                         |
| Module/theme search           | [Text input](https://design-system.service.gov.uk/components/text-input/)                                 | Visible label and a search input; no bespoke search widget.                                                                                                                                     |
| Study programmes              | [Checkboxes](https://design-system.service.gov.uk/components/checkboxes/) in a Fieldset                   | Several choices may apply at once. Small controls are intended for compact filter interfaces.                                                                                                   |
| Prüfungsordnung               | [Radios](https://design-system.service.gov.uk/components/radios/) in a Fieldset                           | Exactly one of three filter settings (all, 2023, 2017). All is the default filter, not a prefilled answer to a personal question.                                                               |
| Search submission             | Button and reset link                                                                                     | Submit an action; navigate to the unfiltered archive.                                                                                                                                           |
| Exam records                  | [Table](https://design-system.service.gov.uk/components/table/)                                           | One table with a caption and scoped headers. Module/date and programme/PO remain directly comparable without repeated headers. A labelled, focusable overflow container supports small screens. |
| Team responsibilities         | [Summary list](https://design-system.service.gov.uk/components/summary-list/)                             | Responsibility is the key; assigned members and contact details are the values. Standard rows align on desktop and stack on mobile.                                                             |
| Optional programme/PO filters | [Details](https://design-system.service.gov.uk/components/details/)                                       | One supplementary section for narrowing results. Never hides search or results; stays open when filters are active.                                                                             |
| Failed enhanced search        | [Inset text](https://design-system.service.gov.uk/components/inset-text/) with an alert role              | Directly relevant service failure, not a field-validation error or an unrelated notification.                                                                                                   |

The [Select guidance](https://design-system.service.gov.uk/components/select/)
recommends considering radios first and treating selects as a last resort for
public-facing services. The short PO select has therefore been replaced by radios.
The [Notification banner guidance](https://design-system.service.gov.uk/components/notification-banner/)
recommends putting information directly related to the current task in the page
content instead; no notification banner or validation-error component is used for
a failed search request. No-results remains ordinary body text.

## Decisions for the preview

### Navigate a service

Source: [Navigate a service](https://design-system.service.gov.uk/patterns/navigate-a-service/).

This is a repeatedly used site with independent tasks, so persistent navigation
fits better than a sequential task list. The preview uses standard Service
navigation markup, an active-page indicator and its mobile menu behaviour.
The site name links to the preview root. The skip link bypasses navigation and
leads to the main content.

Navigation contains only implemented local pages. The footer contact link clearly
names the current website and stays in the same tab: an intentional departure
from the pattern's usual external-link recommendation, avoiding unexpected new
windows.

### Contact a department or service team

Source: [Contact a department or service team](https://design-system.service.gov.uk/patterns/contact-a-department-or-service-team/).

Keep responsibilities and email addresses visible on the team page. Show a
shared role address once after its members, and personal addresses alongside
individual names. A short contact address should not be hidden inside Details. Group elected
members separately from project positions as the current page does.

For the future contact page, order channels consistently based on student needs
and the team's capacity; put social media last. Add opening hours and expected
response times only when the team has confirmed them. Do not invent availability.

### Question pages and validation

Sources: [Question pages](https://design-system.service.gov.uk/patterns/question-pages/)
and [Recover from validation errors](https://design-system.service.gov.uk/patterns/validation/).

The archive is a browsing tool, not a multi-step application. Keep related search
and filter controls on one page rather than applying one-question-per-page to
them. Use visible labels, a fieldset/legend for study-programme checkboxes and
an associated programme hint. Multiple study programmes can be selected; no selection
means all programmes. Prüfungsordnung replaces the earlier invented semester
filter to match the existing archive.

Use an explicit submit button instead of updating results while someone types.
This makes result changes predictable. GET parameters retain the filters in
shareable URLs; ordinary submissions work without JavaScript. HTMX enhances the
same form with result-only updates and history entries. After an enhanced search,
focus moves to the result heading. Module exams remain visible in one captioned table.
Use `hx-swap="outerHTML"` because the response includes the results wrapper. An
inner swap would nest duplicate result IDs on successive searches. Disable HTMX's
DOM history snapshots and reload from the URL on back/forward: form edits made
before submitting must not be restored against the previous search's results.
The server remains the source of truth for restored filters.
If the production archive becomes too long, use linked module pages and pagination
as needed rather than hiding essential results in Details.

No matches is a normal result, with a reset link, not a validation error. A failed
request gets a retry/reload message while retaining the form and previous results.
It must not be presented as a mistake in the user's input.

For future write forms, validate on the server after submission, retain entered
values, prefix the page title with `Fehler:`, focus an error summary and link it to
inline field errors. Disable native HTML5 validation with `novalidate`; do not
validate on blur or introduce required-field asterisks.

## Patterns for later migration work

| Pattern                                                                                                                                                                                                                 | Application and decision                                                                                                                                                                                                   |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Create accounts](https://design-system.service.gov.uk/patterns/create-accounts/)                                                                                                                                       | Let students use public content before authentication is necessary. Preserve actual archive access rules. Separate `Anmelden` from `Konto erstellen` when account journeys are implemented; the preview has no sign-in UI. |
| [Email addresses](https://design-system.service.gov.uk/patterns/email-addresses/), [Passwords](https://design-system.service.gov.uk/patterns/passwords/), [Names](https://design-system.service.gov.uk/patterns/names/) | Apply when migrating account forms. Ask only for needed data, support password managers and paste, and accept real-world names. These inputs are absent from the preview.                                                  |
| [Dates](https://design-system.service.gov.uk/patterns/dates/)                                                                                                                                                           | Choose date entry based on the upload task and known exam date; keep display dates distinct from editable fields. Do not collect dates merely to fill metadata.                                                            |
| [Check answers](https://design-system.service.gov.uk/patterns/check-answers/)                                                                                                                                           | Consider a review step for a multi-step reconstruction upload. Let students change metadata without re-entering other answers. Do not insert a review step into archive search.                                            |
| [Confirmation pages](https://design-system.service.gov.uk/patterns/confirmation-pages/)                                                                                                                                 | After a successful upload, explain whether it is received, awaiting review or published, and what happens next. Only confirm actual persistence; include relevant next links and a receipt/reference if available.         |
| [Page not found](https://design-system.service.gov.uk/patterns/page-not-found-pages/)                                                                                                                                   | Future styled 404: `Seite nicht gefunden`, useful navigation/contact, correct HTTP status, no blame or red warning treatment. Not yet implemented in the preview.                                                          |
| [There is a problem with the service](https://design-system.service.gov.uk/patterns/problem-with-the-service/)                                                                                                          | Unexpected server failures need a clear explanation and useful next step. Preserve answers where possible, and accurately say whether they were saved. A failed enhanced search currently has an inline recovery message.  |
| [Service unavailable](https://design-system.service.gov.uk/patterns/service-unavailable-pages/)                                                                                                                         | Use for deliberate closure, separately from unexpected failure. Give a reopening time only when known, explain the fate of in-progress work, and return an appropriate status. Not implemented in the preview.             |
| [Cookies page](https://design-system.service.gov.uk/patterns/cookies-page/)                                                                                                                                             | Review actual cookies when sessions or optional tracking are introduced. Explain necessary cookies and provide choices for optional ones. The read-only preview has no cookie-consent UI.                                  |

## Patterns without a current user need

Bank details, payment cards, National Insurance numbers, equality information,
postal addresses and phone-number confirmation do not serve these two preview
pages. Username creation and email confirmation belong to later account decisions.
Start pages, suitability checks, interruption pages, step-by-step navigation and
multi-task journeys should only be introduced for a demonstrated task; simple
team lookup and exam browsing should remain directly accessible. Quick-exit
behaviour has no identified use in these journeys.

## Browser inspection and verification

Used Playwright with cached headless Chromium after the desktop browser bridge
was unavailable. No browser setup or project dependencies were required from the
user. Inspected screenshots of the reference and both local pages at 1440px and
390px; also measured layouts at 1024px, 768px and 320px, including expanded filters.
No page-level horizontal overflow was found. At 320px the results table has its
own keyboard-scrollable overflow container.

The earlier page buried results below a long form and repeated table headers for
every module. The revised layout makes results visible alongside desktop filters
and combines them into one table. Team responsibilities now align with members
and contacts in responsive Summary list rows.

Chromium interaction checks passed for repeated HTMX searches, unique result IDs,
result focus, combined filters, reload, back/forward with matching form values,
reset, empty results, a simulated failed request and retry, native GET forms with
JavaScript disabled, the mobile menu and skip link. Following the documentation
shell update, repeated searches, history, mobile navigation and the inert header
search were rechecked; no uncaught page JavaScript errors were observed.

Manual screen-reader and user preference review remain outstanding.

## Altklausuren UI/UX review

Reviewed the rendered archive, not just its templates: default, expanded filters,
selected filters, empty results and failed-search recovery. Compared with the
reference site's shell and GOV.UK guidance for Details, Checkboxes, Radios,
Tables and navigation. The findings below describe the reviewed version; the
follow-up fixes are recorded after them. They were not confirmed WCAG failures.

### Findings

1. **Medium — ambiguous programme state.** `program-hint` always says `Alle`,
   including at `/archive?program=Informatik&version=2023`, where only Informatik
   is checked. The requested short wording should represent a real state or
   action: show it only for an unrestricted selection, or make it an actual
   all-programmes/reset control. Static hint text should not contradict the
   checked options.
2. **Medium — narrow-screen results are clipped.** At 320px, the default results
   table is about 348px wide inside a 290px container. The page itself does not
   overflow and keyboard scrolling works, but programme names and the end of the
   header are initially cut off. Prefer wrapping long programme names before
   relying on a horizontal scrollbar. Retain the standard text size.
3. **Medium — collapsed active filters have no visible status.** Active filters
   are initially expanded after a reload, but users can collapse them. The summary
   then still says only `Filter`; it gives no indication that restrictions remain
   active. Use a concise active count or state in the summary. Do not add another
   explanatory block. This also makes the hidden secondary controls easier to
   discover, consistent with the Details guidance's emphasis on descriptive text.
4. **Low — misleading home breadcrumb.** `Startseite` points to `/`, which redirects
   to `/team`. That is not an archive overview or a real homepage. Remove this
   breadcrumb until there is a meaningful parent, or use the destination's actual
   name. Copying the reference's breadcrumb is not enough to justify its hierarchy.
5. **Low — redundant keyboard stop and generic region label.** The table wrapper
   is tabbable even when it does not overflow. Make it keyboard-focusable when
   scrolling is actually needed. Axe also asks for manual review of
   `aria-labelledby` on the generic results div; a labelled section would express
   the intended region more clearly.

### Follow-up fixes

All five findings were addressed:

- `Alle` is shown only when no programme is selected. JavaScript updates it as
  checkboxes change; server rendering preserves the state without JavaScript.
- The table uses equal-width columns, language-aware hyphenation and wrapping
  rather than clipping long words. At 320px it fits the 290px content area.
- The disclosure shows the number of selected programme options plus a selected
  PO, for example `Filter (2)`. This reflects the current form selection; explicit
  submission still applies that selection to the results.
- Removed the misleading breadcrumb from the shared shell.
- Results use a labelled section. A ResizeObserver makes the table wrapper
  keyboard-focusable only if it overflows, including after HTMX replacement.

Follow-up Chromium checks at 1440, 390 and 320px passed for selection/count
updates, collapsed labels, server-rendered state after reload, table fit and
focusability. Rechecked HTMX replacement/focus and native GET submission with
JavaScript disabled. Axe reported no violations at these widths, including no
remaining generic-region label finding.

### What passed

- Blue header, pale navigation, 1100px container and desktop content alignment
  match the chosen documentation-site reference. The shell is a local adaptation,
  not a prescribed GOV.UK archive pattern.
- Visible input labels, fieldsets/legends, multi-select checkboxes, single-choice
  radios, table caption and scoped headers use the appropriate semantics.
- Results stay visible; one supplementary Details section holds optional filters.
  There is no accordion around the actual exam records.
- Search uses explicit submission with native GET fallback. Repeated HTMX
  searches, unique result IDs, result focus, URL history, reset, combined filters,
  empty results and failed-request recovery passed browser checks.
- Mobile navigation and skip link work by keyboard. The header search remains an
  intentionally nonfunctional visual prototype, separately labelled as Demo.
- No page-level horizontal overflow at 320, 390, 768, 1024 or 1440px, including
  expanded filters. A 720px reflow check approximating a 1440px viewport at 200%
  zoom also passed; this is not a manual browser-zoom or screen-reader audit.
- Axe reported no confirmed WCAG A/AA or best-practice violations at the four
  audited widths (320, 390, 1024, 1440). It left the generic results label for
  manual review, and contrast of clipped content at 320px was inconclusive.

The archive still has fictional records and no file-opening/downloading journey.
That is a documented demo limitation. A real migration needs an obvious module or
file action; a metadata-only results table does not complete the student's task.

## Events & Galerie: pre-implementation review

Reviewed the existing public flow in
`website/app/events/[[...slug]]/page.tsx` and the event/media API response types.
This is an event-album browser, not a calendar or booking journey. Events expose
a title, optional cover and creation timestamp; media expose a title, description,
uploader, MIME type and creation timestamp. A creation timestamp must not be
presented as the date the event took place.

### Guidance and decisions

- [Images](https://design-system.service.gov.uk/styles/images/): photographs are
  the content of this journey and have a clear user purpose. GOV.UK does not
  provide a ready-made gallery component. Use a small local responsive image-grid
  layout with ordinary lists, links and captions. Do not add decorative stock
  photography or fabricate event details. Local illustration fixtures can exercise
  the demo flow but must be clearly identified as examples.
- [Tabs](https://design-system.service.gov.uk/components/tabs/): do not use tabs
  as page navigation or preload all albums. Use named album links, wrapping on
  narrow screens, instead of copying the existing horizontal strip with its hidden
  scrollbar. Do not require users to discover sideways scrolling to find albums.
- Album overview → album → image is a meaningful hierarchy. Keep explicit links
  to the overview and back to the containing album. Preserve the originating
  album page and thumbnail anchor on the return link. Do not repeat a long album
  list above the image on mobile or force a sidebar onto every page.
- [Pagination](https://design-system.service.gov.uk/components/pagination/): use
  numbered pagination below album thumbnails only when content volume or measured
  loading cost justifies it. Do not show pagination for a single page or use
  infinite scrolling. Include the current page in the document title. Reset the
  page when switching albums, redirect obsolete page numbers to the first page,
  and omit previous/next links at the corresponding boundary.
- A dedicated image page is the initial full-size viewer. Show the uncropped
  image with its original aspect ratio, meaningful title/caption, image position
  (for example `Bild 2 von 12`), album return link and labelled previous/next
  links. GOV.UK block pagination can supply the navigation styling; its use for
  image sequences is a local adaptation. Do not wrap the last image to the first
  silently. Normal URLs support sharing, reload, history and JavaScript-free use.
- [Functional images](https://www.w3.org/WAI/tutorials/images/functional/): every
  thumbnail must be a real link with a meaningful accessible name. Use one link
  for thumbnail and visible title rather than two consecutive links to the same
  image. Empty alt is appropriate for a redundant cover beside its linked album
  title; full images need concise descriptions of their visual content. Generic
  `Bild`, filenames and uploader names are not visual descriptions.
- [Modal dialog](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/): a later
  lightbox enhancement would require focus placement/containment, Escape, an
  explicitly labelled close button, inert background and restored thumbnail
  focus. The dedicated page provides the baseline first; a modal is not necessary
  to complete the public journey.
- [Breadcrumbs](https://design-system.service.gov.uk/components/breadcrumbs/):
  add only if they clarify the hierarchy beyond the album navigation. Avoid
  inventing a homepage or duplicating navigation. Any breadcrumbs belong before
  main content so the skip link bypasses them.
- Use responsive previews, explicit image dimensions and lazy loading for
  off-screen thumbnails. Do not lazy-load the primary full-size image. Keep text
  below images rather than overlaying it, preserve visible focus and allow text
  to wrap at 320px. Full-size images must fit the viewport without cropping.
- Keep empty albums distinct from unknown album/image URLs and loading failures.
  An empty album gets a short message and an overview link; unknown resources get
  HTTP 404. Do not represent failed image loading as an empty album.
- Media have MIME types; do not blindly render video or unsupported files as
  images. A photo-only fixture demo must document that narrower scope. Future
  video support needs native controls and appropriate captions/transcripts.

### Implementation acceptance checks

Inspect the rendered overview, album and image pages at 1440, 768, 390 and 320px,
including portrait/landscape images and long titles. Exercise links by keyboard,
skip link, mobile navigation, return-to-thumbnail, first/last image boundaries,
page changes, direct URLs, reload and browser history with JavaScript on and off.
Check image loading, aspect ratios, accessible names, empty albums, 404 responses
and automated accessibility results. This review establishes design decisions;
these checks were subsequently exercised as recorded below.

### Implemented gallery and browser review

Implemented `/community/events`, `/community/events/{album}` and
`/community/events/{album}/{image}` with
server-rendered Go templates and a separate read-only gallery domain package.
After the initial illustration prototype, the user requested actual public
photos. Downloaded six photos each from LAN Party 2025 (45 source images), 2024
(133) and 2023 (50), including portrait/landscape formats, on 16 October 2026.
The source API's titles and descriptions were empty. Added descriptive demo
titles and alt text after inspecting a contact sheet; did not infer identities
or use upload timestamps as event dates. Local 400px and 1200px JPEG previews
total about 2.5MB. Counts explicitly identify the selection and link to complete
source albums. Empty/single-image behavior remains covered by renderer tests.
Production photo
pagination size must be chosen against actual preview sizes/loading performance;
the demo's four-item size is not a production recommendation.

The latest user reference is the [Community section](https://design-system.service.gov.uk/community/).
Community is a top-level navigation item with a concise overview and section
sidebar linking to Events & Galerie and Team. It stays highlighted on descendants;
the sidebar identifies the current page/section. Mobile section links wrap rather
than becoming a horizontal scroller. Team's in-page anchors remain on desktop.
This supersedes the initial full-width gallery layout. `/events/...` URLs redirect
to the Community paths while retaining queries. Album links wrap in a responsive
grid, with an optional cover and visible title/count. Thumbnails and
their text share one native link. Image pages show descriptive alt text,
uncropped content, image position, boundary-aware navigation and an album return
URL preserving page and thumbnail anchor. Page numbers appear in the document
title and pagination rather than lengthening the main heading.

Inspected desktop and mobile screenshots, including the official Community page.
Playwright checks at 1440, 768, 390 and 320px covered Community, the real-photo
album overview, both album pages, the portrait viewer and Team.
Images decoded successfully; no page-level horizontal overflow or automated Axe
violations were found. The complete browse/page/open/return/previous flow passed
with JavaScript enabled and disabled, including direct reload, history, first/last
boundaries and invalid page redirection. Focus and unique results replacement in
archive search were rechecked after adding gallery navigation. Go HTTP tests
cover missing albums/images, album ownership, pagination boundaries, return URLs
and empty/single-image albums. Manual screen-reader review remains outstanding.

### Landing-page adaptation

Inspected the rendered [Design System homepage](https://design-system.service.gov.uk/)
and its HTML structure alongside the current fsInformatik homepage and public
discussion feeds. Adopted its full-width blue masthead, pale updates surface,
topic columns and simple section breaks. The homepage has no section sidebar.
The intro retains the current site's study/events/exams message and Forum/Galerie
links. Added direct routes to archive, gallery, team and Community. A snapshot of
the actual announcement and two forum posts retains publication dates, adapted
plain-text summaries and original `/d/{id}` links. The agenda does not invent
future dates or treat the January 2026 event as upcoming; it identifies the
snapshot and links to the current agenda. Snapshot date is visible.

The brand link now reaches a real homepage rather than redirecting to Team.

The user subsequently requested closer homepage alignment and an event carousel.
Re-read the GOV.UK [pattern index](https://design-system.service.gov.uk/patterns/),
[Navigate a service](https://design-system.service.gov.uk/patterns/navigate-a-service/),
[Start using a service](https://design-system.service.gov.uk/patterns/start-using-a-service/)
and [Layout](https://design-system.service.gov.uk/styles/layout/) guidance before
editing. The site supports multiple repeat tasks rather than a linear transaction:
use top-level section navigation, not a task list or a mandatory start-page gate.
The Design System homepage is a visual reference, not itself a reusable GOV.UK
homepage pattern. Retain concise intro text, a primary Forum link and secondary
Galerie link; use the existing inverse Service navigation variant on the homepage
to join the header, navigation and masthead into one blue composition. Updates
and agenda share a two-thirds text column on a pale blue surface, followed by
three topic columns and simple two-thirds content sections.

For the custom carousel, read the W3C [Carousels tutorial](https://www.w3.org/WAI/tutorials/carousels/),
[functionality guidance](https://www.w3.org/WAI/tutorials/carousels/functionality/)
and [APG Carousel pattern](https://www.w3.org/WAI/ARIA/apg/patterns/carousel/).
The three real album previews occupy the masthead's right-hand third on desktop
and stack below the introduction on narrow screens. Use a labelled region,
named slide groups, native previous/next buttons and a polite status for changes.
Focus stays on the activated control; hidden slides and their links leave the
accessibility/tab order. The carousel changes only on request, with no automatic
rotation or animation. This avoids requiring pause controls or reduced-motion
exceptions. Without JavaScript, the first linked cover and ordinary Galerie link
remain usable; controls are hidden. This is a local component adaptation, not a
GOV.UK-provided carousel. Keep essential navigation outside it so finding other
pages never depends on discovering a slide.

Homepage and carousel follow-up checks passed at 1440, 1024, 768, 390 and 320px:
native keyboard activation, focus retained on previous/next buttons, one exposed
slide link, correct position announcements, valid album links and JavaScript-free
fallback. Axe found no violations and there were no uncaught page errors. Visual
inspection caught full-width mobile button styling squeezing the status; fixed
the carousel controls to 44px wide. Added the requested header login button as a
normal link to the existing site's `/login`, without adding demo auth/session code.

## Sprache und Inhaltsblöcke im Fachschaftskontext

Auf Wunsch wurden am 16. Oktober 2026 öffentlich lesbare Startseiten und
Angebotsbeschreibungen verglichen. Dies ist eine qualitative Stichprobe, keine
repräsentative Sprach- oder Nutzungsstudie:

| Quelle                                                                                                                  | Beobachtung für die Überarbeitung                                                                                                          |
| ----------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| [FSV Informatik / Studienstart an der WH](https://www.w-hs.de/informatik/einstieg-ins-studium/studienstart-informatik/) | Lokale Grundlage: Alle Informatik-Studierenden bilden die Fachschaft; die FSV vertritt sie ehrenamtlich und unterstützt bei Studienfragen. |
| [AStA Westfälische Hochschule](https://www.asta-wh.de/)                                                                 | Konkrete Angebote, „Rund ums Studium“, „Veranstaltungen“, „Kontakt“; Beratungsangebote sind ausdrücklich Leistungen des AStA.              |
| [iFSR TU Dresden](https://ifsr.de/)                                                                                     | Direkte Ansprache, „Unterstützung im Studium“, „Veranstaltungen & Events“, „Mitmachen & Engagement“, „Neuigkeiten“.                        |
| [FSMI am KIT](https://www.fsmi.org/)                                                                                    | „Von Studierenden für Studierende“, „Termine“, „Aktuelles“, „Mach mit!“ sowie konkrete Klausur- und Kontaktangebote.                       |
| [Fachschaft Informatik Regensburg](https://fsi.ur.de/)                                                                  | Studienalltag als Ausgangspunkt: Altklausuren finden, Fragen stellen, andere Studierende kennenlernen, Anliegen einbringen.                |
| [AStA TU Braunschweig](https://astatubs.de/)                                                                            | „Service und Angebote“, „Beratung“, „Neuigkeiten“, „Termine“, „Aktiv werden!“; sachliche Angebotsnamen mit direkter Ansprache.             |
| [AStA Uni Mainz](https://asta.uni-mainz.de/)                                                                            | Interessenvertretung, praktische Unterstützung, studentisches Engagement und Veranstaltungen als Begegnungsräume.                          |

### Abgeleitete Entscheidungen

- Durchgängig „du/dein“ für einzelne Studierende, „wir“ für die Fachschaftsvertretung.
  „Studierende“ ist verständlich und inklusiv; Fachbegriffe wie FSV werden nicht
  ohne Erklärung zur Einstiegshürde. Übernommen wird die direkte Ansprache,
  nicht der jeweils unterschiedliche Slang der Quellen.
- Der Einstieg erklärt, wer hier spricht, für wen die Seite ist und was sie
  anbietet. Er nennt die WH und Gelsenkirchen statt eines austauschbaren Slogans.
- „Neues aus der Fachschaft“ ersetzt „Ankündigungen“. „Veranstaltungen“ ersetzt
  „Agenda“. Die Seite verlinkt zu Terminen, statt einen leeren Kalender und
  technische Erklärungen zum Datenstand in den Vordergrund zu stellen.
- Drei Hauptangebote: „Altklausuren“, „Austausch im Studium“ und „Deine
  Ansprechpersonen“. Jeweils ein kurzer Aufgabentext und ein konkreter Link.
  Die Galerie bleibt über Masthead, Karussell und Community erreichbar.
- „Studierende helfen Studierenden“ erklärt den Zweck des Forums. Originale
  Beitragstitel und Datumsangaben bleiben erhalten; redaktionelle Kurzfassungen
  dürfen sprachlich angepasst werden.
- „Mitmachen“ ersetzt den generischen Schlussblock „Community“. Einstiege sind
  Veranstaltungsorganisation, eigene Ideen und Kontakt zum Team. Nicht gewählte
  Projektstellen sind im eigenen Team belegt.
- „Galerie“ bezeichnet vergangene Veranstaltungen/Fotos; „Veranstaltungen“ führt
  zu Ankündigungen mit Terminen. Vergangene LAN-Partys werden nicht als kommende
  Events angekündigt. Der gewünschte Navigationsbegriff „Community“ bleibt.
- Keine fremden Leistungen übernehmen: Rechtsberatung, Sozialberatung,
  Ticketservice, Sprechstundenzeiten oder bestimmte Gremienmandate werden nicht
  allein deshalb versprochen, weil andere Fachschaften/ASten sie anbieten. Die
  allgemeine Kontaktadresse stammt von der eigenen öffentlichen Kontaktseite.
- Der Aufbau folgt der GOV.UK-Referenz: kurzer Einstieg mit Handlung,
  hervorgehobene Neuigkeiten, drei Angebote, vertiefende Abschnitte. Die Sprache
  beschreibt Aufgaben und nächste Schritte, keine Design-System-Begriffe.

Die überarbeiteten Seiten Startseite, Community und Galerie wurden bei 1440,
1024, 768, 390 und 320px erneut in Chromium geprüft. Kein horizontaler
Seitenüberlauf und keine bestätigten Axe-Verstöße; das Karussell blieb bedienbar.
Desktop- und Mobilansichten wurden zusätzlich anhand von Screenshots betrachtet.
Die bestehenden Go-Tests bestanden nach der Überarbeitung.

## Carousel sizing and homepage alignment review

The subsequent rendered-layout review found that the old desktop-only column
breakpoint made the carousel expand to 708px at a 768px viewport. The masthead
became approximately 943px tall. The independently spaced Forum button and
Galerie link did not form a coherent action group, while primary study links
appeared more than 1,100px down the desktop page.

- Use a responsive masthead grid with vertically centred columns from 641px;
  retain the two-thirds/one-third composition above 768px. Cap the carousel at
  400px, or 320px when stacked on phones. Preserve the image's 3:2 presentation.
- Use the GOV.UK button group for Forum and Galerie, with consistent gaps and
  no trailing child margins. Keep the inverse start-button styling.
- Keep the linked album title under the photo; shorten the live position text
  to “Album n von 3” rather than repeating that title. Retain manual controls,
  44px targets, focus retention and the no-JavaScript fallback.
- Move the three principal offers immediately below the masthead, before news.
  Use intrinsic shared grid rows (subgrid) to align descriptions and links without
  fixed text heights. Stack each complete offer on smaller screens. Rename the
  long “Deine Ansprechpersonen” heading to “Kontakt zur Fachschaft” so it fits
  narrow desktop columns.

Final Chromium checks covered 1440, 1024, 769, 768, 640, 390 and 320px. There was
no page or content overflow, desktop descriptions/actions aligned, carousel
keyboard activation retained focus and updated its live position, and Axe found
no confirmed violations. The first album remains linked with controls hidden
without JavaScript. Screenshots were inspected at desktop and phone widths.
At 768px the masthead measured approximately 398px instead of 943px.

## Sign-in and typography review

Reviewed the official [accounts](https://design-system.service.gov.uk/patterns/create-accounts/),
[passwords](https://design-system.service.gov.uk/patterns/passwords/),
[question pages](https://design-system.service.gov.uk/patterns/question-pages/),
[password input](https://design-system.service.gov.uk/components/password-input/)
and [error summary](https://design-system.service.gov.uk/components/error-summary/)
guidance before creating `/login`. Inspected the rendered password-input example
and the existing application's login implementation.

- Use a left-aligned two-thirds content column with no empty sidebar, decorative
  card or carousel. Email and password form one sign-in task. Keep recovery and
  account creation separate from the submit action, with explicit external links.
- Ask for the full email address rather than splitting it into local-part/domain
  controls. After feedback, remove the introductory account/ZA explanation,
  email hint and all preview wording from the sign-in page.
  Use `autocomplete="username"` with `type="email"` and `current-password`.
- Use the standard password component with German show/hide labels and live
  announcements. Allow paste and password managers; do not impose new password
  creation rules on sign-in or repopulate passwords after submission.
- Validate server-side with matching summary/inline messages, linked fields,
  focus on the summary and a `Fehler:` document-title prefix. Empty/invalid email
  and missing password return HTTP 422. Complete input shows “Anmeldung derzeit
  nicht verfügbar” with a link to the existing login rather than an invented
  credential failure or success. No authentication,
  session, credential persistence or forwarding is introduced. Responses use
  `Cache-Control: no-store`; form body parsing is bounded.
- Public content remains accessible without signing in. The login entry is an
  optional account task, not a mandatory start-page gate. No progress indicator
  is needed for this standalone entry page.

Following the user's shell review request, re-read the official
[layout](https://design-system.service.gov.uk/styles/layout/) and
[back-link](https://design-system.service.gov.uk/components/back-link/) guidance.
The sign-in shell keeps the brand and public section navigation; it omits the
inert search, redundant header login button and preview footer text. Place a
specifically labelled “Zur Startseite” link before main (fixed destination,
not browser history), followed by the standard two-thirds form column and a
`govuk-heading-l` title. Keep the email/password fields together, the standard
left-aligned submit action, recovery link and a distinct “Noch kein Konto?”
section containing only the account-creation link. No introductory paragraphs or
decorative card. The shell is a local composition of documented components,
not an official standalone GOV.UK sign-in-page pattern.

The user subsequently requested the reference typeface. Reviewed the official
[typeface guidance](https://design-system.service.gov.uk/styles/typeface/) and
inspected computed styles on the Design System homepage: its search input uses
GDS Transport at 19px/25px. The local prototype now embeds the regular and bold
WOFF2 fonts from the pinned GOV.UK Frontend package, and the asset refresh script
copies them alongside the CSS. The custom carousel caption uses the same family.
Official guidance requires an alternative typeface for publicly available services
outside service.gov.uk; this is a local reference exploration, not a deployment
decision. Search placeholder and icon now use dark text, with full opacity; on
narrow screens search has a full-width row beneath brand/login so its label fits.

Final sign-in shell checks at 1440, 1024, 768, 390 and 320px verified the initial
page, validation errors, summary focus/field links, German password toggle and
unavailable response focus. No horizontal overflow; without JavaScript, password
entry and server validation still work. Desktop/phone screenshots were inspected.
Axe reported only its `region` advisory for the back link outside main, the known
placement issue documented by GOV.UK's back-link guidance. No other violations
were reported on the initial page. Earlier font-regression checks covered home,
archive, team, Community and gallery at the same widths without Axe violations
or page overflow. Go tests and vet passed.

## Allowed account email domains

Reviewed the [email-address](https://design-system.service.gov.uk/patterns/email-addresses/)
and [email-confirmation](https://design-system.service.gov.uk/patterns/confirm-an-email-address/)
patterns following the user's account-eligibility clarification. The demo login
retains a single full-address field with the concise hint “Nur @studmail.w-hs.de
oder @fsv-wh.de”. This is actionable field guidance, not introductory page copy.
The hint and any inline error are associated through `aria-describedby`.

`internal/account` owns the reusable server-side email policy: valid full mailbox,
exact case-insensitive domain match, no subdomains or lookalike suffixes. The
handler trims surrounding whitespace and preserves the local part. Disallowed
domains receive a specific linked validation error without checking account
existence. Tests cover both domains, casing, whitespace at the HTTP boundary,
display names, quoted local parts and misleading domain suffixes.

This implements eligibility in the isolated demo. Production signup/sign-in,
email changes and recovery need the same rule when integrated. Signup must
confirm access to the allowed mailbox; recovery responses must not disclose
whether an account exists. No confirmation emails or production auth changes
are introduced by the form preview.

## Password recovery page

Added `/forgot` after reviewing the official
[password recovery guidance](https://design-system.service.gov.uk/patterns/passwords/#help-users-who-forget-their-password)
and [service unavailable pattern](https://design-system.service.gov.uk/patterns/service-unavailable-pages/).
The existing email, validation and account-shell decisions apply. One email field
with `autocomplete="email"`, the same allowed-domain hint and a “Link anfordern”
button are sufficient; no security questions, password reminder, password field
or repeated email entry. The back link above main returns to `/login`, whose
“Passwort vergessen?” link now opens the local form.

The handler reuses `account.EmailError`, preserves invalid email input, links the
inline error and summary, and prefixes the error document title. Requests are
bounded and responses use `no-store`. The demo has no mail delivery: eligible
addresses receive the same HTTP 503 unavailable page with a link to the existing
recovery service, without account lookup, token generation or a claimed email
delivery. This is deliberately not a fake success confirmation. Once a real
recovery service is integrated, use a neutral response regardless of account
existence and time-limited, single-use reset links as described by the pattern.

Verified the login → recovery → validation → unavailable → login journey at
1440, 1024, 768, 390 and 320px, plus submission without JavaScript. Summary focus,
field links, retained invalid email, domain errors, status codes and absence of
horizontal overflow passed. Axe found no issues beyond the already documented
back-link `region` advisory. Desktop/phone screenshots were captured and the
phone rendering inspected. Go tests and vet passed.

## Fachschaft as a main navigation section

At the user's request, add “Fachschaft” before Community and Altklausuren in
the standard Service navigation. The overview at `/fachschaft` introduces the
student representation and links to Team and Sitzungen und Protokolle.
Its section navigation contains “Über uns”, “Team” and “Sitzungen und
Protokolle”, following the same documentation-site shell as Community. Active
ancestors use `aria-current="true"`; the current section page uses `page`.

Team retains its existing `/team` URL and now belongs to Fachschaft instead of
Community, with a general contact address above the responsibility lists.
Following further feedback, the Team entry contains nested links to Vorstand,
Referate and Projektstellen when viewing that page. The separate Kontakt section
and its nested link were removed at the user's request. Vorstand and the
elected departments are separate groups; “Referate” names the areas of work,
rather than using the person label “Referenten”. These links replace
the separate page-contents navigation. Use “und” or separate labels instead of
ampersands in authored UI labels. The page title and section-navigation entry
were finally renamed to “Mitglieder”; the existing `/team` URL is retained.
Community keeps overview and gallery navigation, with cross-links for engagement.
The homepage engagement link now leads to Fachschaft. Meetings/minutes live at
`/fachschaft/sitzungen`; no dates, meeting rules, decisions or documents are
invented. The initial empty states were subsequently replaced with the supplied
document archive described below. A contact route remains available for topics.

### Supplied invitations and minutes

Imported the user's `FSV Informatik_20261016_2218.zip` from Downloads. Examined PDF
text for document headings, meeting dates and source inconsistencies, plus the
rendered latest invitation. The catalog covers 49 source paths, 48 unique PDF
byte streams and 31 session entries in years 2020–2026. A content-identical copy
of the 6 March 2020 minutes appears under two filenames; one linked PDF retains
both source paths. Original PDF bytes are unchanged. The import script checks
complete archive coverage; HTTP tests check every PDF's size and SHA-256 hash.

Use Summary list rows keyed by date/session, with invitation and minutes links
as values. Include PDF type and size, and unique accessible link names. Year links
are ordinary server-rendered URLs, with the newest supplied year as default and
`aria-current` for the selected year. Missing document types are simply absent;
an invitation alone is not proof the meeting occurred. No search, expandable
cards or JavaScript filter is needed for seven years and 31 entries.

Specific source issues are retained in document notes:

- April 2026 invitations name 2025 in the meeting sentence; issue dates and
  corresponding minutes establish the 2026 index date. The second invitation
  also has a first-session heading.
- The invitation filed in 2025 names 23 January 2024 but was issued 7 January 2025. Group it by the source folder, label the date as ambiguous and display
  both original year claims rather than inventing a corrected session date.
- Three 2022 minutes have 2022 headings and 2021 opening sentences. The index
  follows the heading, with a visible discrepancy note.
- The March 2020 minutes contain a conflicting January heading. The file named
  `Protokoll 4_1` describes the third session on 13 November 2020 and is grouped
  by that content rather than falsely paired with the December invitation.
- The April 2024 second-session minutes call it the first session in their
  opening text; the date agrees with the invitation and the note keeps the
  numbering discrepancy visible.

The latest supplied minutes schedule a further meeting for 15 October 2026,
already past on the 16 October snapshot date. No upcoming date can be established
from these documents, so the next-session empty state remains.

## Forum discussion listing

The user clarified that the requested page is the actual `/discussions` listing,
titled “Forum”, rather than an introductory landing page. Reviewed the official
[Navigate a service](https://design-system.service.gov.uk/patterns/navigate-a-service/),
[Lists](https://design-system.service.gov.uk/styles/lists/),
[Text input](https://design-system.service.gov.uk/components/text-input/) and
[Pagination](https://design-system.service.gov.uk/components/pagination/) guidance.
Inspected the rendered existing public forum, its Next.js implementation, the
public discussions API and the rendered GOV.UK input example. There is no official
GOV.UK forum pattern: this is a local composition of documented components.

- Keep the page under Community, with Forum selected in section navigation and
  Community marked as the active top-level ancestor. Canonical route `/discussions`
  matches the user's requested terminology and existing application.
- Use a short introduction, an explicit “Beitrag verfassen” link and a real GET
  search form. Visible labels identify search and post type; no placeholder-only
  inputs. Both filters apply on submission and preserve state in the URL.
- After feedback, present “Beitrag verfassen” as the primary GOV.UK button and
  “Suchen” as a secondary button. The create action remains a native link to
  the source editor, with GOV.UK's button-link keyboard handling. Post titles
  and the filter-reset navigation remain links.
- The post-type select keeps its native HTML semantics and popup, but normalises
  the closed control with `appearance: none`, square corners and a local SVG
  arrow to avoid browser-specific rounded styling. Forced-colour mode restores
  native appearance. GOV.UK focus and border treatments remain intact.
- Present a semantic list of posts with linked heading, brief editorial summary,
  plain-text type and compact original author/date/comment metadata. Titles are
  links, rather than making the whole row a JavaScript click target. Use thin
  separators; omit decorative cards, avatars, menus and nonfunctional vote buttons.
- Include all five publicly returned posts in a dated embedded snapshot. Keep
  original publication chronology (not event date or latest edit) and metadata.
  Search includes original body text and tags as well as displayed summaries.
- Post-type options are All, Discussions, News and Events, labelled in German.
  An empty state offers a concrete next step; reset clears both filters. Unknown
  type parameters fall back to all types. Search text is escaped by html/template.
- Per pagination guidance, do not add pagination to a five-item list merely to
  demonstrate the component. A future larger dataset should use numbered pages,
  filter-before-pagination and a page number in the document title.
- Homepage forum actions now open the local listing; event/news links select the
  appropriate type. Individual threads and the existing `/d/new` editor remain
  on the source application, with the destination explained on the list page.
  Public reading is not gated behind the disconnected demo login.

Checked Chromium at 1440, 1024, 960, 768, 390 and 320px for initial list, search,
combined filters, empty state, reset and browser back. Inspected desktop/phone
screenshots. No horizontal page overflow; Axe found no confirmed violations on
the initial listing. Search and filtering also work without JavaScript. Unit and
HTTP tests cover body/tag search, multiple terms, combining filters, ordering,
metadata, empty results, escaping and active Community navigation.

## Discussion details

`/d/{id}` now renders all five source posts locally, including all eight public
comments. Revisited [Paragraphs](https://design-system.service.gov.uk/styles/paragraphs/)
and [Back link](https://design-system.service.gov.uk/components/back-link/) guidance.
This is a composition of typography, links, lists, inset text and Summary lists,
rather than an official forum-detail pattern.

- Keep a single post-title H1 and compact author, publication and edit metadata.
  Preserve original wording, line breaks and paragraphs; render source bullet
  lists and quoted comment text semantically. Decode stored HTML entities once,
  then let templates escape text. Only HTTP(S) URLs are linkified. No raw source
  HTML is rendered, and no general Markdown dependency is needed for this snapshot.
- Show event date/time and room in a Summary list. Dates reflect the source,
  without describing the January 2026 event as upcoming.
- Comments form a newest-first ordered list with author headings, publication
  dates, edit notices and stable local permalinks. The primary “Kommentieren”
  button opens the source post, with an adjacent explanation of the destination.
- A back link before main returns to the URL-derived search and type filter;
  it works without JavaScript. Forum becomes an active ancestor (`aria-current=true`)
  in section navigation, with Community active in the main navigation.
- Homepage post links now open local details. Creating and commenting still open
  the existing application. This supersedes the external-thread-link decision
  recorded in the earlier listing review.
- User-requested navigation wording is now “Fachschaftsvertretung” in the main and
  section navigation. Long post titles wrap to prevent narrow-screen overflow.

Browser review: inspected desktop and phone screenshots; tested all five posts at
1440, 1024, 768, 390 and 320px, list-to-detail-to-filtered-list journeys and a
JavaScript-disabled journey. Axe on the LaTeX thread reports only the documented
GOV.UK back-link `region` advisory. Go tests cover complete comment counts,
chronology, lookup/404, event metadata, safe text/link rendering and filter return.

## Creating and editing forum posts

Reviewed [Textarea](https://design-system.service.gov.uk/components/textarea/),
[Recover from validation errors](https://design-system.service.gov.uk/patterns/validation/)
and [Details](https://design-system.service.gov.uk/components/details/) guidance,
alongside the existing application's shared editor and API validation rules.

- One cohesive form serves `/d/new` and `/d/{id}/edit`. Keep one page heading,
  visible labels above each field, short hints and an appropriately sized
  12-row plain-text textarea. Preserve copy/paste, links and line breaks; omit
  formatting controls unsupported by the read view. Edit preloads the original
  values, including topic labels and event details.
- Three explicit radios expose the post types without requiring users to open a
  select. Optional event date, time and room are in one Details disclosure; it is
  open for events, errors or retained event values. This keeps non-event forms
  concise and works without JavaScript. The first conditional-radio experiment
  produced an Axe `aria-expanded` advisory on the native radio after GOV.UK
  initialisation; the final native Details version avoids that issue.
- Match the source title/body length rules and use Unicode character counts.
  Optional event dates accept day.month.year; times use 24-hour hours:minutes.
  A supplied time requires a date. No new client-side validation or truncating
  maxlength attributes: use `novalidate`, validate on submission, retain all
  answers, prefix document titles with “Fehler:”, and provide linked summary and
  inline errors. GOV.UK moves focus to the error summary.
- Use a primary “Beitrag veröffentlichen” / “Änderungen speichern” button with
  an “Abbrechen” link. Back/cancel goes to the list for create and source detail
  for edit. The detail's secondary edit button and list's primary create button
  now lead locally.
- Valid input is not treated as a saved post: return HTTP 503 with a focused
  notification, an explicit not-saved explanation, intact form values and a link
  to the existing create/edit screen. No account, role, ownership, persistent
  write or forwarding is simulated. These forms remain local design exploration.

Browser review covered create, discussion edit and event edit at 1440, 1024, 768,
390 and 320px, error-summary focus, invalid date recovery, valid/unavailable
submission, field retention, no overflow and the JavaScript-disabled flow.
Desktop/phone screenshots were inspected. Axe reports only the documented
back-link `region` advisory, with no confirmed violations on these editor states.
Go tests cover validation, Unicode limits, input retention/escaping, prefilling,
bounded requests, unknown IDs and the not-saved response.

Latest editor refinement: removed the complete “Beitragsart” radio group following
user feedback. New posts use the discussion type; existing posts keep their source
type, derived on the server rather than from a hidden or submitted field. Optional
event details now appear only when editing an existing event. This supersedes the
type-selection decision above; title, body and topics remain the general form.

## Review before adopting the design

### Rich-text integration review

#### Final implementation: GDS UI, existing editing engine

Final live-reference comparison used the official textarea and secondary-button
examples in Chromium, not just palette documentation. The official textarea's
focused computed styles are 3px `#ffdd00` outline, 2px `#0b0c0c` border and a 2px
inset black shadow, 0px radius, GDS Transport 19px/23.75px and 5px padding. The
local editing surface now matches these values. Yellow is only a focus state.
Custom styles use GOV.UK CSS colour properties; toolbar background is generated
from the Sass black tint-95 palette value. The iframe also loads the same GOV.UK
stylesheet so font definitions and link/error/focus tokens stay consistent.

Native-control browser checks passed at 1440, 1024, 768, 390 and 320px for matching
focus/typography, flat horizontal wrapping toolbar, link insertion/edit/removal,
invalid-URL recovery, error-summary focus, keyboard navigation, formatting retention
and no-JavaScript fallback. Inspected final desktop/mobile screenshots. Scoped
Axe found only the documented back-link region advisory, including the link-error
state. No TinyMCE toolbar/status bar is rendered. The removed topics field is
absent. Later feedback moved the URL example into the input placeholder while
retaining its visible label and linking any validation error directly to the input.

User direction clarified that only the UI should change. TinyMCE remains the
editing engine, with its own toolbar, status bar, dialogs and context menu
disabled. The HTML template now renders a real GOV.UK button-group toolbar and
labelled inline link fields. This supersedes the engine-replacement recommendation
and TinyMCE-skin experiments below.

The actual GDS list, ordered-list, link, undo and redo icons and MIT licence are
vendored at revision `c053cf1c198580723c8768f938cb6146dcd122e8`. Bold, italic and
quote are local SVGs. Button surfaces are flat except on hover/focus, following
GDS's editor toolbar, with 44px controls and accessible German names. Arrow keys,
Home/End, Alt+F10, pressed-state announcements and disabled undo/redo states are
handled explicitly. Link insertion/edit/removal uses native GOV.UK fields rather
than an editor dialog or browser prompt; selection is preserved while editing.

Investigated the reported rounded focus border in computed styles: TinyMCE's
`.tox-edit-area::before` applied a blue 2px border with 4px radius. Disabled its
focus highlight and removed that pseudo-border; focus/blur events now toggle a
square yellow/black outline on the editing container. Iframe fonts were already
GDS Transport, but their line height was 28.5px versus the native textarea's
23.75px. Changed to 19px/1.25 with 5px margin, using the existing regular and bold
WOFF2s. Compared toolbar colours with the compiled GOV.UK 6.6 stylesheet and
replaced legacy `#f3f2f1`/`#b1b4b6` values with `#f3f3f3`/`#cecece`. Corrected the
button-group's mobile column layout to horizontal wrapping. The themes field was
removed from create/edit per feedback; existing source tags remain internal.

#### Follow-up: an actual GOV.UK editor UI

After the user rejected TinyMCE's overall visual fit, researched concrete GOV.UK
implementations rather than more CSS reskinning. Found
[GDS' Govspeak visual editor](https://github.com/alphagov/govspeak-visual-editor),
an MIT-licensed ProseMirror-based demonstration. Its toolbar actually uses
`govuk-button govuk-button--secondary`, `govuk-button-group`, `govuk-select`
and repository SVG assets. `scss/menu.scss` removes button shadows outside hover,
giving a flatter editing toolbar than the site's submit-action buttons. The
editable area uses the GOV.UK black border/yellow focus treatment and publishing
typography. This is a closer visual reference than an independently skinned editor.

Cloned to the approved temporary directory, installed with lifecycle scripts
disabled and ran the upstream demo separately on port 8091. Inspected its rendered
desktop screenshot and source (`lib/plugins/menu/buttons.js`, `menu-plugin-view.js`,
`scss/menu.scss`, `scss/prosemirror-overrides.scss`). No replacement of the site's
editor was made during this research.

Important suitability findings:

- The README explicitly calls it a demonstration; it is not a supported GOV.UK
  Design System rich-text component. Its package currently declares GOV.UK
  Frontend 4.8 rather than the prototype site's 6.6.
- Content output is Govspeak (a Markdown variant), including publishing-specific
  constructs. The source schema currently includes links as its only text mark;
  bold/italic would need adding. A forum should omit CMS callouts, address blocks,
  attachments and other specialist publishing features.
- Link creation uses `window.prompt`, not a GOV.UK-styled accessible form. A local
  integration should replace this with a labelled inline link form and preserve
  editor selection/focus.
- Scoped Axe on the upstream editor reported `aria-input-field-name`, `image-alt`
  and `select-name`. The image rule includes toolbar icons and example content.
  These are findings in the upstream demo, not a claim about every deployment.
- The demo's explicit 750px minimum editor width produces an 822px-wide page at a
  390px viewport. This example layout needs replacement, not copying unchanged.
- The old MOJ/HMCTS editor remains archived for accessibility reasons. A 2020
  TinyMCE/GDS React gist exists, but is only a historical custom integration, not
  a maintained official skin or accessibility endorsement.

Revised recommendation: use the GDS editor as the concrete visual/interaction
reference for a small ProseMirror editing surface with real GOV.UK controls and
its appropriately attributed SVGs. Keep the existing safe HTML/form boundary;
do not import the full Govspeak publishing feature set. The native toolbar would
need accessible names/pressed states, robust keyboard navigation, responsive
layout and manual screen-reader testing. Replacement awaits direction following
this research.

Reviewed the [GOV.UK rich-text discussion](https://github.com/alphagov/govuk-design-system/issues/5847),
the [archived MOJ component](https://design-patterns.service.justice.gov.uk/archive/rich-text-editor/),
and TinyMCE's current accessibility/navigation/conformance and licensing guidance.
GOV.UK has no supported rich-text component; MOJ withdrew its own implementation
as insufficiently accessible. This is a local TinyMCE integration, not a claim
of GOV.UK endorsement or a completed accessibility audit.

User feedback rejected the default TinyMCE icon styling and then clarified that
icons should remain. Final controls use local monochrome SVGs with square buttons,
the site's black/grey palette and yellow/black focus. German tooltips and accessible
names identify commands; toggle buttons expose active state. Toolbar keyboard
navigation remains managed by TinyMCE. The visible “Für Hilfe ⌥0 drücken” prompt
is disabled using `help_accessibility: false`, while Help remains in the toolbar.

Only safe basic formatting is available. Reuse the existing Go bluemonday
dependency for a strict server HTML allowlist, and keep plain/rich input handling
explicit rather than guessing whether text is HTML. Preserve the field label,
description, validation and error-summary journey. The fallback warns about
formatting loss, retains link destinations and never exposes HTML as editable
source. Valid submissions still report unavailable rather than claiming to save.

Final Chromium checks passed at 1440, 1024, 768, 390 and 320px: visible enhanced
editor with the fallback hidden, icon controls, bold/list/quote commands, rich
submission and retained formatting, safe rendered preview, error-summary focus,
toolbar keyboard entry (Alt+F10), link/help dialogs and Escape. JavaScript-disabled
and blocked-editor-script flows retain a usable plain textarea. Desktop and phone
screenshots were inspected; no horizontal overflow. Axe reported only the known
GOV.UK back-link `region` advisory. A missing accessible name on TinyMCE's branding
link is supplied locally. Manual VoiceOver/NVDA testing remains outstanding.

Sources: [TinyMCE accessibility options](https://www.tiny.cloud/docs/tinymce/latest/accessibility/),
[conformance report information](https://www.tiny.cloud/docs/tinymce/latest/accessibility-conformance/),
and [self-hosted licensing](https://www.tiny.cloud/docs/tinymce/latest/license-key/).

- Compare desktop and narrow-screen navigation, team grouping and archive labels
  with the existing site. Check at 200% zoom and for horizontal overflow.
- Exercise skip link, mobile menu, filters and results by keyboard;
  verify visible focus and meaningful screen-reader headings/announcements.
- Verify ordinary GET submissions with JavaScript disabled, enhanced submissions,
  reset links, browser back/forward and history restoration with populated filters.
- Simulate a failed search request and check that previous results remain usable
  and the recovery message is announced.
- Confirm with students that programme checkboxes, PO selection and module
  grouping are useful; GOV.UK components alone do not establish usability.
# Account creation review

## Return-link review

Live-reference placement correction: inspected `/components/back-link/` DOM and
bounding boxes. Breadcrumbs belong inside the right-hand content column, directly
above its main/h1, beside the sidebar. The split-pane container now wraps a sidebar
and content column; breadcrumbs are the main element's preceding sibling within
that column. Verified alignment with h1 at 1440, 768, 390 and 320px, no overflow.

Latest navigation decision: breadcrumbs replace the previously reviewed top-level
home/back links on all non-home pages. Following the live
[Breadcrumbs guidance](https://design-system.service.gov.uk/components/breadcrumbs/),
they sit before main and include only ancestors, starting at Startseite. Examples:
Startseite → Fachschaft for members/meetings, Startseite → Community → Forum for
post details/create, with the post as an additional parent when editing. Gallery
images include their album, retaining the image-return page/anchor. No breadcrumbs
appear on home, and no back component appears alongside them. Native GOV.UK markup
supplies chevrons, typography and focus; long ancestor titles can wrap. Chromium
checks passed at 1440, 390 and 320px for account, team, meetings and forum pages.
The return-link review below records the preceding iteration.

### Nested side navigation

Compared the rendered navigation and stylesheet on `/components/back-link/` and
its `/history/` child. Nested lists have 20px left padding, 10px top margin and no
guide border. List items use 5px vertical padding and 5px bottom margin; links use
2px vertical padding and 16px/20px type. A 4px brand-colour border marks the whole
active top-level list item (including its children), with -14px left margin and
10px left padding. Only the link with `aria-current="page"` is bold. Applied these
visual rules to both Community and Fachschaft navigation. Verified create/edit and
team pages at 1440, 768, 390 and 320px, without overflow; inspected mobile screenshot.

Compared local account links with the live [Back link example](https://design-system.service.gov.uk/components/back-link/default/)
and [Back link guidance](https://design-system.service.gov.uk/components/back-link/),
alongside [Navigate a service](https://design-system.service.gov.uk/patterns/navigate-a-service/).
Computed styles match the reference: 16px/20px GDS Transport, 15px vertical
margins, 14px left padding, black underlined text and CSS chevron. Keyboard focus
uses the standard yellow background and black underline. Links are before `main`,
so the skip link bypasses them, and ordinary hrefs work without JavaScript.

The login page's fixed homepage destination is not a return to the previous page;
it now uses an ordinary GOV.UK link labelled “Zur Startseite” without a back chevron.
Register/recovery retain the back component, labelled “Zurück zur Anmeldung”.
Forum detail/create/edit already use explicit destination labels and retain filtered
forum URLs where supplied. No breadcrumbs are combined with back links. GOV.UK
explicitly documents the out-of-landmark automated `region` advisory; moving these
links into main solely to suppress it would undermine skip-link behaviour.

## Comment-section review

GOV.UK provides no dedicated forum-comment component or pattern. Reviewed the
existing section against its [headings](https://design-system.service.gov.uk/styles/headings/),
[links](https://design-system.service.gov.uk/styles/links/) and
[button](https://design-system.service.gov.uk/components/button/) guidance.
Retained a section heading with count, newest-first list, author headings,
machine-readable dates and descriptive comment permalinks, standard body text,
lists and quotations. No extra controls or panels are needed. Comment separators
now use the current `--govuk-border-colour`. The external commenting destination
is a clearly labelled ordinary link, in the same tab, replacing the button and its
redundant explanatory sentence. Snapshot date remains separate.

Re-reviewed the live GOV.UK guidance for [Create accounts](https://design-system.service.gov.uk/patterns/create-accounts/),
[Email addresses](https://design-system.service.gov.uk/patterns/email-addresses/),
[Passwords](https://design-system.service.gov.uk/patterns/passwords/) and
[Password input](https://design-system.service.gov.uk/components/password-input/)
for the local `/register` page.

- Use “Konto erstellen” consistently and a dedicated form, separate from login.
  Public browsing remains available without an account.
- Collect name (the existing API's public display name), one complete email address
  and one password. Explain public name visibility and email purpose. No repeat
  email/password field, CAPTCHA or security questions.
- Use GOV.UK input, password reveal, hints, field errors and linked error summary;
  `autocomplete="name"`, `email` and `new-password`, with paste allowed and no
  client-side password truncation. Preserve name/email on errors, never passwords.
- Registration accepts only the exact `studmail.w-hs.de` domain on the server.
  Login and recovery continue accepting `studmail.w-hs.de` and `fsv-wh.de`.
- Account creation now follows the GOV.UK password length guidance: at least eight
  Unicode characters, no password-specific maximum and no character-composition
  rules. Whitespace is preserved. The overall 64 KiB form-body limit still applies.
  Name length remains 2–64. The production API's separate 8–72 password policy is
  unchanged. Common-password checking is omitted at the user's request.
- No account or email verification is simulated: valid input returns 503 and
  explains that no account was created or email sent. All responses are no-store.

Chromium checks passed at 1440, 768, 390 and 320px: login-to-create navigation,
password show/hide, error-summary focus and field links, exact-domain rejection,
name/email retention and blank password after submission, unavailable response,
and no horizontal page overflow. Reviewed the 390px screenshot. No-JavaScript
submission also works. Axe in the submitted state reported only the previously
documented back-link region advisory. Go tests and vet passed. The shared footer
no longer renders the OGL mark or licence sentence.
