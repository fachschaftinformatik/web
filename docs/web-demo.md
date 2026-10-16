# Server-rendered Go preview

Run the local preview with:

```sh
go run ./cmd/web-demo
```

For a read-only snapshot of real user display data, pass the production-copy
directory explicitly:

```sh
go run ./cmd/web-demo -data .dev-data/production-2026-10-14
```

The SQLite connection uses `mode=ro` and selects only user IDs, names, email
addresses, roles, active status, email verification status and programme names. It excludes deleted users
and never selects passwords, reset/verification tokens or production sessions.
The connection closes after loading the snapshot. Source files are not embedded
in the binary; the default remains fictional user fixtures. The supplied snapshot
has 362 non-deleted accounts and no team/responsibility/meeting tables.

## Local administration demo

Sign in at <http://127.0.0.1:8090/login> with **admin / admin**. This is an explicit
demo-only credential exception; ordinary login/recovery/registration retain their
email-domain policy and unavailable-service behaviour. Successful demo login
rotates an opaque, server-side session (eight-hour lifetime). The header then
offers Verwaltung and a POST Abmelden action. Administration pages and all
mutations require authentication; POST forms use per-session CSRF tokens.
Responses containing session data use `Cache-Control: no-store`. Logout invalidates
the session on the server. The cookie is HttpOnly, SameSite=Lax and Secure on TLS.

- `/verwaltung`: index and contextual navigation.
- `/verwaltung/benutzer`: case-insensitive name/email search, 20 accounts per page,
  linked account details, lowercase email display and empty-result states.
  Status shows Bestätigt/Nicht bestätigt using the independent `verified`
  column. Inactive accounts also show Konto inaktiv; details separate Kontostatus
  from Status rather than treating activation as email confirmation. Numbered
  GOV.UK pagination includes first/last/current/adjacent pages, ellipses and
  directional links. The document title includes the current page and total;
  invalid pages redirect to the first page while retaining the search query.
- `/verwaltung/benutzer/{id}`: summary list and Passwort zurücksetzen action.
  JavaScript opens a native modal dialog with the target email address, initial
  focus on its heading, Escape/cancel closure and focus restoration. Without
  JavaScript, the same action opens a normal confirmation page at
  `/verwaltung/benutzer/{id}/passwort-zuruecksetzen`. Confirming POSTs to that URL
  and redirects to an explicit demo notice. No email, token or password change is
  produced; this previews the future reset-link email operation.
- `/verwaltung/mitglieder`: add/edit public member profiles independently of
  login accounts. Board responsibility text is editable on existing board profiles.
- `/verwaltung/referate` and `/verwaltung/projektstellen`: add/edit titles,
  shared email addresses and member assignments with small checkboxes. Removing
  a checked member removes the assignment, not the person.
- `/verwaltung/sitzungen`: existing meetings and document links. Add an invitation
  or minutes PDF to a new or existing meeting using its date and number.
- `/verwaltung/sitzungen/neu` and `/verwaltung/sitzungen/{id}`: a document form
  with GOV.UK date input, document-type radios and translated improved file upload.
  PDF uploads are limited to 8 MiB, with basic PDF header/end-marker checks and a
  32 MiB total process-local document store. Invalid metadata and review/edit
  navigation retain valid uploads; the owner can preview a draft PDF.
- `/verwaltung/sitzungen/{id}/pruefen`: check-answers summary, contextual edit
  links and explicit publication. Conflicting date/number identities are rejected;
  invitations and minutes share a meeting. Repeat publication is idempotent.
- `/documents/{id}`: published PDFs are public; unpublished files are owner-only.

Member/responsibility edits update `/team`; published documents update the public
meeting archive and upcoming-meeting section. These changes are stored only in
memory and reset on restart. The database/bucket copy remains a source snapshot.
Team and meeting operations live in their domain packages, with HTTP handlers
mapping forms, sessions and responses. Public account deletion and actual email
delivery remain future features.

Verification includes Go integration tests for access control, login/session
rotation/logout, CSRF, read-only snapshots, public edit visibility, upload
validation/retention, draft ownership, publication and reset confirmation.
Go tests pass with the race detector, and Go vet passes. Playwright Chromium and
Axe checks cover administration pages at 1440, 1100, 768, 390 and 320px; numbered
pagination and modal keyboard behaviour were checked at 1440, 768, 390 and 320px.
The publication review was checked at 1440, 390 and 320px. Browser journeys cover
member edits, department/project assignments, retained uploads through validation
and review edits, publication, no-JavaScript protocol publication and reset
confirmation, and logout. A before/after SHA-256 comparison confirmed that the
source SQLite file stayed unchanged during the mutation journeys.

- Team: <http://127.0.0.1:8090/team>
- Archive: <http://127.0.0.1:8090/archive>
- Startseite: <http://127.0.0.1:8090/>
- Community: <http://127.0.0.1:8090/community>
- Events & Galerie: <http://127.0.0.1:8090/community/events>

The archive uses eight fictional uploads across six modules. Search, combined filters, empty results,
shareable URLs, and module grouping work. Direct PDF links serve a clearly labelled
example file, generated by `python3 tools/update-web-demo-archive.py`. No real exam
PDFs or exam uploads are implemented. Administration sessions, document uploads
and the optional read-only user snapshot are described above. Team contacts mirror the public
team configuration, copied for preview rather than a new source of truth.

Events & Galerie provides an album overview, paginated thumbnails and individual
image pages with previous/next links, image position and return-to-thumbnail
links. It uses six real photos from each of the public LAN Party 2025, 2024 and
2023 albums on ik.fsv-wh.de, fetched on 16 October 2026. These are a selection,
not the complete albums: counts and links to the originals make this explicit.
The 400px thumbnails and 1200px viewer previews are embedded locally (about 2.5MB
total); the demo makes no runtime requests to the source for images. Source titles
and descriptions were empty, so descriptive titles and alt text were authored for
this demo after visually inspecting the photos. Gallery media IDs and album titles
remain those of the source. The original API identifies Francesco La Camera as
the uploader; that is not an assertion of photographic authorship or a licence.
Photos are source-site content, separate from GOV.UK's asset licences.
The gallery supports these JPEG fixtures only, with no live API/storage connection
or video viewer. All navigation works without JavaScript. Unknown
albums/images return 404; invalid album page numbers redirect to the first page.
Existing `/events/...` links redirect to `/community/events/...`, preserving queries.

The top-level navigation is “Fachschaft · Community · Altklausuren”. Fachschaft
(`/fachschaft`) introduces the student representation; its sidebar links to
“Über uns”, “Mitglieder” (`/team`) and “Sitzungen und Protokolle”
(`/fachschaft/sitzungen`). Team now belongs to Fachschaft, with nested links to
Vorstand, Referate and Projektstellen on the team page. The meetings page
groups invitations and minutes by session, with year links (`?jahr=2026`, etc.).
The supplied `FSV Informatik_20261016_2218.zip` contains 49 PDFs, imported as 48
unique, byte-preserved documents for 31 meetings across 2020–2026. A byte-identical
duplicate has both original source paths recorded against one document. Source
filenames, sizes, SHA-256 checksums and date-conflict notes are in
`internal/webdemo/meetings/catalog.json`. PDFs are embedded under `assets/meetings/`.
Ambiguous source dates are labelled instead of silently corrected; the 2025
invitation has no asserted date. No future meeting is invented. Re-import with:

```sh
python3 tools/import-web-demo-meetings.py '/path/to/FSV Informatik_20261016_2218.zip'
```

Community is inspired by the Design System's `/community/` page and groups
exchange, events and gallery. Its sidebar links to overview and gallery and
marks the current one. On mobile the short
section list wraps above the content. The homepage follows the Design System's
continuous blue header/navigation/masthead, three main offers directly below it
and a pale blue updates area. A manual event-album carousel sits alongside the
introduction from 641px and below it on smaller screens. It is capped at 400px
(320px on phones), with a linked album title and a short position indicator.
Native previous/next buttons keep focus, hidden
slides are removed from the tab order and changes have a polite status. It does
not auto-rotate. Without JavaScript the first linked cover and Galerie link work,
and the controls stay hidden. Its introduction was adapted to the Fachschaft
context, with a snapshot of the public announcement and two
forum posts. Summaries are adapted, original dates are shown and links open local
detail pages. The former Agenda block is now “Veranstaltungen”: it describes
student events and links to current announcements without pretending the snapshot
contains upcoming dates. No auto-refresh is implemented.

Copy and content groups were revised after comparing the WH's own information,
Fachschaften in Dresden, Karlsruhe and Regensburg, and ASten at the WH, TU
Braunschweig and Uni Mainz. The homepage now uses “Neues aus der Fachschaft”,
“Veranstaltungen”, the three offers “Altklausuren”, “Austausch im Studium”, “Kontakt
zur Fachschaft”, followed by “Studierende helfen Studierenden” and “Mitmachen”.
Community distinguishes event announcements, past-event photos and engagement.
The album browser is labelled “Galerie”. See the language research in
`govuk-patterns.md` for sources and rationale.

## Forum

`/discussions` is the Forum listing, under Community. Homepage and Community
forum links now reach this local page; news/event links preselect the relevant
post type. The snapshot was retrieved on 16 October 2026 from
`https://ik.fsv-wh.de/api/v1/discussions?limit=30` and contains the five public
posts (three discussions, one news item, one event). Original titles, authors,
publication dates, tags, bodies and comment counts are retained in
`internal/webdemo/forum/snapshot.json`; summaries are editorial adaptations.

GET parameters `q` and `type` provide working server-rendered search across titles,
bodies, summaries and tags, combined with post-type filtering. Results are newest
first. Search uses all terms, case-insensitively, over the complete snapshot.
Controls retain the URL-derived values, and reset returns to `/discussions`.
There is no runtime API dependency or pagination for five items. Each post links
to a local `/d/{id}` detail page; “Beitrag verfassen” opens the local `/d/new`
form and “Beitrag bearbeiten” opens `/d/{id}/edit`. Publishing, commenting and
voting remain on the existing application.
Forum editing remains a preview; the local administration session described above
does not connect it to production or add forum persistence.

Detail pages include the original body, post type, author, publication and edit
dates, topic labels and all eight public comments across the five posts. Comment
permalinks are local anchors; comments are newest first. `forum/threads.json`
contains public detail and comment API snapshots retrieved on 16 October 2026
from `/api/v1/discussions/{id}` and `/api/v1/discussions/{id}/comments` on the source
site. The event includes its source date, time and location. Source entities are
decoded once, then escaped by templates; paragraphs, bullet lists, quotations and
HTTP(S) links use semantic markup without trusting source HTML. The list passes
search/type parameters to detail links, and the explicit back link reconstructs
the filtered list URL without JavaScript. Unknown IDs return 404.

The main navigation and section-navigation heading now say
“Fachschaftsvertretung”, following user feedback; the overview remains `/fachschaft`.

Create and edit share a server-rendered form with title and body. At the user's
request, the post-type selector and topics field were removed. New posts use
the discussion type; editing retains the source type server-side, ignoring posted
type values. Existing event edits also offer optional date/time/location. Editing loads the
source values and decodes stored entities for plain-text editing. Event details
use a native Details disclosure, open for events or retained event values. Both
forms validate on POST (HTTP 422, linked summary/inline errors); valid submissions
return HTTP 503 with an explicit not-saved message, retained inputs and a link to
the corresponding existing editor. Title/body limits match the existing API
(3–255 and 10–10,000 Unicode characters). Forms are no-store, bounded to 128 KiB
and never persist or forward submitted content. They are public exploration
screens, not a connected authenticated editing workflow. Unknown edit IDs are 404.

## Rich-text editor

Create and edit progressively enhance the plain-text body with self-hosted
TinyMCE 8.10.0, German translations 26.9.28 and the lists plugin.
Refresh with `python3 tools/update-web-demo-editor.py`; editor dependencies and
licences are embedded in `assets/tinymce/`. The local GPL configuration uses no
cloud service. TinyMCE's toolbar, dialogs, status bar and extra focus highlight
are disabled. The UI uses native GOV.UK secondary buttons and input fields,
following the GDS Govspeak editor's flat toolbar. The list/link/undo/redo SVGs and
MIT licence are vendored from GDS revision `c053cf1c198580723c8768f938cb6146dcd122e8`
in `assets/editor-icons/`; bold, italic and quote use local SVGs. No ProseMirror
dependency is added. All buttons are 44px, with German accessible names and
pressed/disabled state. Arrow/Home/End keys navigate the toolbar; Alt+F10 enters it.
Ctrl/Cmd+K opens a labelled inline link form, with editing/removal, safe URL checks,
Enter to apply and Escape to return to the toolbar. No browser prompts or TinyMCE
link dialogs appear, and no keyboard-help prompt is displayed.

The toolbar uses GOV.UK 6.6's secondary-button background `#f3f3f3` and active/hover
grey `#cecece`, rather than older warm-grey palette values. GDS Transport regular
and bold are loaded inside the editing iframe; 19px text, 1.25 line height and 5px
margin match the native GOV.UK textarea. The editor's explicit focus state uses
square black borders and a 3px yellow outline, suppressing TinyMCE's rounded blue
inner highlight. Toolbar controls wrap horizontally on narrow screens.

The ordinary textarea stays usable without JavaScript or if the editor fails to
load. Once initialised, the rich editor posts `rich_body` instead of `body`.
Server sanitisation allows only paragraphs, breaks, emphasis, lists, quotations
and safe links. Validation counts visible text rather than HTML tags. Error-summary
links focus the editor, which carries the field's accessible name and error state.
Rich formatting survives validation and the unavailable-save response, with a
sanitised rendered preview. A plain-text fallback retains text and link destinations;
its hint explains that posting this version removes formatting. Detail rendering
also supports an explicit `BodyFormat=html` through the same sanitisation boundary;
the source snapshot remains plain text. Persistence is still unconnected.

## Domain-first structure

```text
cmd/web-demo/                         Standalone entrypoint
internal/webdemo/
  account/                           Allowed account-email policy
  archive/                           Exam records and search behavior
  forum/                             Public post snapshot and search/filter rules
  gallery/                           Example albums, images and domain lookups
  home/                              Public landing-page content snapshot
  meetings/                          Curated meeting/document catalog and provenance
  team/                              Members and responsibilities
  handler/http/                      Thin HTTP handlers and presentation data
    templates/                       Layout, pages, archive result fragment
    assets/                          Embedded CSS, JavaScript and sample images
    styles/                          GOV.UK Sass configuration
```

The domain packages have no HTTP or template dependencies. Handlers parse inputs,
call domain functions, and render HTML. A production migration would reuse the
existing persistence/application logic behind these domain operations. This demo
does not introduce repository interfaces or a generic service framework.

`html/template` escapes output. Normal GET forms return complete pages. HTMX
requests return only the result list using the same search operation. History
restoration requests receive complete pages. Search URLs are pushed to browser
history. Exam dates, study programmes and Prüfungsordnung remain directly visible
in a single results table with a caption and scoped headers.
HTMX replaces the complete results wrapper. Back/forward navigation reloads the
URL instead of restoring a DOM snapshot, keeping form values and results aligned.

GOV.UK Frontend 6.6.0 and HTMX 2.0.8 are vendored under `assets/`, with their
licences. GDS Transport WOFF2 files are locally embedded for the requested visual
prototype, with no external font requests. All runtime assets and
templates embed into the binary; no Node or template-generation step is required.
To refresh the compiled assets with curl, tar and Dart Sass installed, run:

```sh
sh tools/update-web-demo-assets.sh
```

The visual reference is <https://design-system.service.gov.uk/get-started/>:
a blue header with a visual-only search field, pale navigation, a 1100px container,
left section/filter sidebar and large page title. Breadcrumbs are omitted until
there is a meaningful parent page. The header search
has no form or endpoint; the archive search is functional. The team uses Summary
lists to map responsibilities to members/contact. A shared header button links to
`/login`, which supports the local admin/admin session described above.
Normal account submissions accept only complete addresses at `studmail.w-hs.de` or `fsv-wh.de`
using the shared `account` policy, validates missing passwords, clears passwords on every response,
and reports that sign-in is unavailable on complete submission. It does not create
production sessions or forward credentials. Account creation links to the local registration form;
password recovery links to the local `/forgot` form. It uses the same domain
policy and account-page shell with a “Zur Anmeldung” link. Invalid input returns
HTTP 422 with linked summary/inline errors. Eligible input returns HTTP 503 with
an unavailable page and a link to the existing recovery service. No email is
sent, account looked up or reset token generated; no success is claimed.
The login shell omits search and the header login button, uses
a two-thirds form column and a visible demo-credentials hint.
Optional programme/PO filters use one Details section, open when active, with
checkboxes and radios. Results use
one table. A short footer notice identifies the preview. There are no government
logos or Crown branding. The [official typeface guidance](https://design-system.service.gov.uk/styles/typeface/)
requires an alternative font for publicly available non-service.gov.uk sites;
this remains a local visual prototype. Local CSS provides
the documentation shell, request indicator, table overflow and email wrapping.
The gallery sits within Community's split-pane shell, with a small local
responsive grid and standard links/pagination. Images retain their aspect ratio;
off-screen thumbnails lazy-load, while the main viewer image loads normally.
Search supports multiple study programmes and a Prüfungsordnung filter; enhanced
search runs on explicit submission. See the [GOV.UK pattern review](govuk-patterns.md)
for sources, adaptations, migration decisions and outstanding interaction checks.

This standalone demo is separate from production routes and container builds.
# Latest prototype refinements

- Archive layout now follows a search-and-filter model rather than a flat
  table. `/archive` has its own full-width layout with an introduction, submit
  action, search, a responsive filter panel and direct PDF links in a result table. Both filter
  groups, plus Abschluss (Bachelor/Master), use small checkboxes without an “Alle” hint or option; an empty
  selection leaves that group unfiltered.
  `version` can repeat in query parameters; empty legacy `version=` still means
  all versions. Informatik und Design is
  available as a programme filter. PDF links match the meetings' document-name
  and “(PDF, size)” presentation; the shared example PDF is labelled as such.
  Submission links point
  to `https://ik.fsv-wh.de/archive`. No demo PDF/upload integration is implied.
  See `govuk-patterns.md` for research and rationale.
  The result table has Modul and Rekos columns, one row per module. The newest
  matching exam has a direct PDF link; modules with multiple matching uploads
  also link to “Alle N Rekos ansehen” at `/archive/{moduleID}`. That page lists
  matching uploads (Datum, Studiengang, Abschluss, Prüfungsordnung and Dokumente columns), with direct PDFs and a
  filter-preserving Altklausuren breadcrumb. There is no individual exam page.
  Explicit module IDs keep grouping independent of display titles. Existing
  example records are Bachelor records; Algorithmen und Datenstrukturen has
  three uploads across two PO versions to exercise the navigation and filters.
  The module page has no introductory paragraph or submit action. Its table
  stacks with labelled fields on mobile; the overview stays a two-column table.

- `/contact` follows the rendered GOV.UK Design System contact page: a single
  two-thirds content column, ancestor breadcrumb, plain headings and linked
  bullet lists. It contains email contacts, Instagram, Discord, WhatsApp and
  the office address; the footer and Fachschaft overview link to it. Social
  links come from `https://linktr.ee/fs_informatik`; office and role email
  addresses are corroborated by
  `https://www.w-hs.de/informatik/gremien/fachschaften/`. The Discord invitation
  uses the Fachschaft's Linktree rather than the different WH-page invitation.
  General email and street address come from the existing contact page.
  Chromium checks passed at 1440, 768, 390 and 320px, including breadcrumbs,
  footer navigation, overflow and Axe; desktop and 320px screenshots reviewed.
- Current Fachschaft sidebar has one visible heading, “Die Fachschaft”, with
  Über uns, Mitglieder, Wahlen, Sitzungen und Protokolle, and Satzung und
  Ordnungen. Mitglieder retains its contextual nested links. This supersedes
  the earlier two-group arrangement described below.

- Meeting archive reviewed against the GOV.UK Table and Summary list guidance
  (`https://design-system.service.gov.uk/components/table/` and
  `https://design-system.service.gov.uk/components/summary-list/`). Repeated
  meeting records now use a table with a year caption, column headers “Sitzung”
  and “Dokumente”, and row headers. Source discrepancy notes remain in the
  catalogue but are not rendered. All seven years passed Chromium overflow
  checks at 1440, 768, 390 and 320px; the 320px layout was visually reviewed and
  Axe reported no violations on the checked page. Handler tests passed.

- Fachschaft includes `/fachschaft/wahlen` and `/fachschaft/ordnungen`, linked
  from the overview and the “Unsere Arbeit” navigation group, with standard
  breadcrumbs. The iFSR Dresden site informed the information architecture only.
  WH sources: `https://www.w-hs.de/amtsblatt/` (student-body statute dated
  23 May 2025 in AM 20/2025; election regulations dated 19 October 2022 in AM
  29/2022) and `https://www.stupa-wh.de/ausschuesse`. PDFs remain at their official
  URLs. No upcoming election dates, candidate lists, or local Fachschaft statute
  were supplied; the pages do not invent these. The Fachschaft's own statute can
  be added to the regulations page when its document is available. Chromium at
  1440, 390 and 320px passed layout, breadcrumb/current-link and Axe checks;
  official PDF links returned 200. Demo Go tests passed.

- Community now has an illustrated “Komm ins Gespräch” section with Forum and
  ideas entries (including the reference's proposal illustration), a two-paragraph
  introduction, an events section with two photo/text rows using the reference's
  `catchup-groupshot.png` and `workshop.png`, and a participation section explaining how to
  bring ideas and contribute through Projektstellen. These reference assets are from
  `alphagov/govuk-design-system` revision
  `a1bfd60d1e0fcc2de52a1d357b934c3098ec62a6`, with upstream MIT licence retained
  in `assets/community-images/GDS-LICENCE.txt`. Refresh them with
  `python3 tools/update-web-demo-community.py`. Decorative images have empty alt
  text; text links provide destinations. Layout is two columns on larger screens
  and one column on phones. Chromium checks at 1440, 768, 390 and 320px passed
  without overflow; Axe found no violations in the reviewed page.

- Navigation labels use “Fachschaft”; discussion details omit the topics line.
- The editor supports normal text and level 2 and 3 headings, including sanitised
  previews and plain-text fallback. The post title remains the level 1 heading.
- The shared footer temporarily uses translated GOV.UK Design System footer text
  and original external destinations and attribution at the user's
  request. These are prototype placeholders to replace with the site's own content.
  Government Crown branding, the OGL logo and the licence sentence are omitted.
- `/register` is the local “Konto erstellen” page, linked from `/login`. It validates
  name, eligible email domain, study programme and password server-side. The eight
  programme options are a local snapshot of the public `/api/v1/programs` endpoint;
  selection is optional and retained on errors. Name, email and password are required.
  Like the other account
  pages it has no connected authentication: valid submissions return 503 with an
  explicit not-created/not-sent explanation; invalid submissions return 422.
  Passwords are never rendered back, and responses use `Cache-Control: no-store`.
