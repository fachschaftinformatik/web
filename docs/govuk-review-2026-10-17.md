# GOV.UK design-system review — 17 October 2026

Reviewed the merged implementation at `62d1d0a` against the GOV.UK
[styles](https://design-system.service.gov.uk/styles/),
[components](https://design-system.service.gov.uk/components/) and
[patterns](https://design-system.service.gov.uk/patterns/) catalogues.

## Scope and evidence

### Implementation follow-up

Findings 1, 2, 4, 5, 6 and 8 have been addressed on this branch:
local service information replaces the GOV.UK footer content, the non-functional
global search has been removed, document failures use shared HTML recovery pages,
comment errors retain input and reopen the correct edit page, validation titles
receive an error prefix, and the forum call to action is a navigation link.
The event cover upload now uses the enhanced drop-zone markup.

Finding 7 has both code and local editorial changes: decorative thumbnails have
empty alternatives with descriptive link text; full-image alternatives retain
curated descriptions; importer publisher/date strings fall back to image titles.
Thirty imported descriptions were rewritten after visual inspection, with original
provenance retained in the ignored curation record. These content changes are in
the local preview database and must be transferred separately from Git.

Finding 3 is explicitly deferred at the user's request. The documented GOV.UK
restriction is not presented as a separately established font-licence violation.

The minimal service information is available at `/impressum`, `/datenschutz` and
`/barrierefreiheit`. The chair comes from the same member assignments as `/team`,
not a hard-coded historical name. The user confirmed AStA provides hosting and
email. The text describes the implemented session cookie, accounts, public and
private profile behaviour, archive submissions and external links. It does not
copy the unrelated analytics/cookie claims in the AStA website's own policy.
Provider-specific logging retention and the formal allocation of data-protection
responsibility/contact should be confirmed with AStA when finalising the text.

Follow-up verification passed:

- Race tests for `internal/web` and `internal/content/media`, `go vet ./...`
  and the application build.
- Chromium/axe checks on 13 page states, including three 404 responses: no
  reported violations. Seven routes checked at 320, 768 and 1280 pixels:
  no horizontal document overflow.
- Error-summary and linked-field focus for new and edited comments, automatic
  reopening of invalid comment edits, event validation titles/focus, and the
  translated improved cover-upload control.
- Thirty local image alternatives checked against the curation record; seven
  eligible albums remain, and the pre-existing foreign-key violation count is
  unchanged at 15. No original image bytes or Instagram captions were changed.

Evidence: `.dev-data/govuk-review/verify-fixes.cjs`, `fixes-report.json`,
`photo-alternatives.json` and desktop/mobile screenshots. The current preview is
served at `http://127.0.0.1:18099/` by `.dev-data/govuk-review/web`.

Sources for the draft service information:

- [HG NRW, §§ 53 and 56](https://recht.nrw.de/lrgv/gesetz/04082026-hochschulgesetz-hg/)
- [DDG, § 5](https://www.gesetze-im-internet.de/ddg/__5.html)
- [Published Fachschaft address](https://github.com/fachschaftinformatik)
- [AStA contact and technical-service context](https://www.asta-wh.de/impressum/)
- [AStA data-protection contact](https://www.asta-wh.de/datenschutz/)

The original review observations below describe the pre-fix implementation.

- Chromium and axe-core checks on 22 public page states and 8 authenticated
  page states: home, community, forum, news, events and pagination, event detail,
  albums, archive, Fachschaft content, contact, sitemap, authentication,
  administration, editors and a discussion with commenting.
- Layout overflow checks at 320, 768 and 1280 CSS pixels on home, forum, events,
  albums, archive and registration. No horizontal document overflow found.
- Interaction checks for global search, event validation, comment validation,
  cover upload initialisation and out-of-range pagination.
- Public content used the curated local preview; authenticated checks used a
  separate disposable database and local administrator fixture.
- Local evidence: `.dev-data/govuk-review/browser-review.cjs` and
  `.dev-data/govuk-review/browser-report.json`. These are ignored development
  files. The report's immediate post-navigation focus observation was followed
  up after module initialisation: event error-summary focus **does work**.
- No axe violations on the 28 successful page responses in this sample. Both
  intentional 404 responses failed document-structure checks. Automated checks
  do not establish screen-reader usability or complete accessibility conformance.

## Findings, in priority order

### 1. High: footer identifies the wrong service and responsible organisation

**Location:** `internal/web/templates/layout.html:114–119`.

The accessibility and privacy links lead to statements for the GOV.UK Design
System website. The footer says this Fachschaft service was developed by the
GOV.UK Design System Team. These are visible service claims, rather than library
attribution. Users cannot obtain this service's accessibility or privacy
information from those links.

**Recommendation:** use Fachschaft-specific destinations and ownership wording;
retain dependency attribution separately where appropriate.

**Reference:** [GOV.UK footer](https://design-system.service.gov.uk/components/footer/)
and [using Frontend without GOV.UK branding](https://frontend.design-system.service.gov.uk/using-govuk-frontend-without-govuk-branding/).

### 2. High: global search presents a task users cannot complete

**Location:** `internal/web/templates/layout.html:23–26` and
`internal/web/assets/site.js`.

Entering `grillen` and pressing Enter on the home page produces no request,
navigation or results. The input has no form owner. Its label and search landmark
promise a working website search on most pages.

**Recommendation:** connect it to a GET search with an explicit submit control
and useful result/empty states, or remove the control until that task is supported.

**Reference:** [text input](https://design-system.service.gov.uk/components/text-input/),
[button](https://design-system.service.gov.uk/components/button/) and
[navigate a service](https://design-system.service.gov.uk/patterns/navigate-a-service/).

### 3. High: GDS Transport is used on a non-GOV.UK service

**Location:** `internal/web/styles/govuk.scss:1–6`, generated
`internal/web/assets/govuk.css`, and `tools/assets.sh:19–20`.

The Sass configuration uses the default font; Chromium confirms paragraphs use
loaded GDS Transport font faces. GOV.UK explicitly directs publicly available
services outside `service.gov.uk` to use an alternative typeface.

**Recommendation:** configure an appropriate system font in the Sass build,
regenerate the CSS and stop copying unused Transport font assets. Include rich
text editor content in the typography check.

**Reference:** [typeface](https://design-system.service.gov.uk/styles/typeface/).

### 4. High: failure responses lose the service page structure and recovery links

**Location:** `internal/web/handler.go:279–320`, plus `http.NotFound` and
`http.Error` calls in individual handlers.

`/missing-review-page` and `/community/veranstaltungen/999` return a bare 404
without a document title, language, main landmark or heading. Each produces five
axe violations. Server errors also use plain-text responses. Users lose the
navigation and contextual recovery guidance available on successful pages.

**Recommendation:** render shared HTML error pages with the correct HTTP status,
German heading, appropriate explanation and useful recovery/contact links.
Apply this to document requests while preserving image/download response semantics.

**Reference:** [page not found](https://design-system.service.gov.uk/patterns/page-not-found-pages/)
and [service unavailable](https://design-system.service.gov.uk/patterns/service-unavailable-pages/).

### 5. Medium: comment validation bypasses the established recovery pattern

**Location:** `internal/web/templates/thread.html:26–31` and comment handlers.

New/edit comment forms use native `required` validation without `novalidate`.
An empty submission produces the browser's English “Please fill out this field”
message in the tested browser. Submitting an empty comment directly to the server
returns HTTP 422 with only plain text. Neither path supplies the linked error
summary and inline error used elsewhere in the service.

**Recommendation:** use consistent server-side validation, preserve entered text,
render a linked summary and field error in the discussion, prefix the document
title with `Fehler:`, and focus the summary after initialisation. Cover editing
as well as creation.

**Reference:** [recover from validation errors](https://design-system.service.gov.uk/patterns/validation/),
[error summary](https://design-system.service.gov.uk/components/error-summary/) and
[textarea](https://design-system.service.gov.uk/components/textarea/).

### 6. Medium: event/editor validation does not identify errors in the page title

**Location:** page construction in `internal/web/events.go` and
`internal/web/editor.go`; title rendering in `internal/web/templates/layout.html:7`.

Submitting an empty event correctly returns an error summary and moves focus to
it once Frontend has initialised. The document title nevertheless remains
`Veranstaltung erstellen · fsInformatik`, without the error prefix recommended by
the validation pattern. Other form paths should be checked for the same omission.

**Recommendation:** set an error-prefixed document title whenever field validation
fails, retaining the normal page heading.

**Reference:** [recover from validation errors](https://design-system.service.gov.uk/patterns/validation/).

### 7. Medium: imported photo descriptions identify the publisher, not the image

**Location:** local imported media descriptions and their presentation in
`internal/web/templates/images.html`.

Thirty images in eligible public albums have descriptions such as
`Photo by Fachschaft Informatik - WH on June 25, 2026.` This records provenance,
but does not communicate the image content. Imported captions are valuable
context, but do not automatically provide a useful alternative for each photo.

**Recommendation:** curate concise descriptions for informative photos; avoid
repeating adjacent link text for decorative previews. Keep provenance separately.
This requires editorial changes to the imported content as well as checking
template fallback behaviour.

**Reference:** [images and alt text](https://design-system.service.gov.uk/styles/images/).

### 8. Low: the home-page forum link uses transaction-start presentation

**Location:** `internal/web/templates/home.html:10`.

“Zum Forum” is a start button with an arrow and button role, although its action
is browsing a discussion list rather than starting a service transaction.

**Recommendation:** use a prominent navigation link; reserve start-button
presentation for an actual start-of-task journey.

**Reference:** [button](https://design-system.service.gov.uk/components/button/)
and [start using a service](https://design-system.service.gov.uk/patterns/start-using-a-service/).

## Working patterns to retain

- German document language, skip link, main landmark, one H1 on sampled success
  pages, labelled navigation and current-page state.
- Responsive layout without sampled document overflow; restrained event covers
  and separate photo albums. Gallery cards and forum voting are service-specific
  patterns, not GOV.UK components requiring visual imitation.
- Bounded event pagination, labelled current page, previous/next links and filter
  preservation. Requesting event page 999 resolves to the final available page.
- Existing form-group errors, linked summaries and retained editor values;
  event error-summary focus works after JavaScript module initialisation.
- Password visibility controls and explicit autocomplete values on account forms.
- Archive review/change steps and summary lists, plus contextual administration
  navigation rather than an unnecessary multi-session task-list pattern.
- File upload translations are configured. The event cover input initialises,
  but does not show the improved drop-zone/button UI used by newer upload markup;
  align its markup with the maintained component when updating that form.

## Suggested implementation order

1. Correct footer ownership/destinations, remove or implement global search,
   configure non-GOV.UK typography.
2. Introduce shared document error pages and consistent comment/error-title
   handling. Verify status codes, retained input and keyboard focus.
3. Curate image alternatives and simplify the forum navigation call to action.
4. Complete keyboard and manual screen-reader review of gallery viewing,
   archive filtering, rich-text editing and error recovery. Existing notes in
   `docs/govuk-patterns.md` remain useful historical context; this review does
   not treat their earlier implementation claims as fresh verification.
