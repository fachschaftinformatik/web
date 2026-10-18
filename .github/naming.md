# Canonical naming

Status: canonical contract for the URL and naming migration. See the [URL tree](urls.md) and
[proxy redirect plan](url-redirects.md).

URLs and code identifiers use English. The UI remains German, with one German
name for each concept. Apply the vocabulary throughout the product, rather than
translating only the navigation or changing only route registration.

## Vocabulary

| Concept / URL | English code vocabulary | German UI vocabulary |
| --- | --- | --- |
| `/about` | `About` | Über uns |
| `/team` | `Team` | Team |
| `/meetings` | `Meeting` | Sitzung / Sitzungen |
| `/elections` | `Election` | Wahl / Wahlen |
| `/regulations` | `Regulation` | Ordnung / Ordnungen |
| `/contact` | `Contact` | Kontakt |
| `/legal` | `Legal` | Impressum |
| `/privacy` | `Privacy` | Datenschutz |
| `/accessibility` | `Accessibility` | Barrierefreiheit |
| `/sitemap` | `Sitemap` | Seitenübersicht |
| `/community` | `Community` | Community |
| `/posts` discussion listing | `Post`, `Posts` | Forum |
| A post | `Post` | Beitrag / Beiträge |
| `/news` | `News` | Ankündigungen |
| `/events` | `Event` | Veranstaltung / Veranstaltungen |
| `/media` | `Media` | Fotos |
| A media album | `Album` | Album / Alben |
| An individual media item | `Item`, qualified as `MediaItem` where needed | Foto / Fotos |
| An actual photograph | `Photo` | Foto / Fotos |
| An actual video | `Video` | Video / Videos |
| `/exams` | `Exam`, `Exams` | Altklausur / Altklausuren |
| An exam module | `Module` | Modul / Module |
| A downloadable document | `Document` | Dokument / Dokumente |
| `/exams/submissions` | `Submission` | Einreichung / Einreichungen |
| A new version of an exam document | `Revision` | Revision / Revisionen |
| `/exams/settings` | `ExamSettings` | Einstellungen für Altklausuren |
| `/users/{user}` | `User`, `Profile` | Benutzerprofil |
| `/admin` | `Admin` | Verwaltung |
| `/admin/settings` | `OrganizationSettings` | Stammdaten |
| `/admin/members` | `Member` | Mitglied / Mitglieder |
| `/admin/roles` | `Role`, qualified as `MemberRole` where needed | Funktion / Funktionen |
| `/admin/users` | `User`, `Account` | Benutzerkonto / Benutzerkonten |
| `/images/...` image bytes | `Image` | Bild / Bilder when a visible label is needed |

"Forum" is the German name of the discussion listing, not an alternative path
or a reason to retain `forum` as the post-domain package name. News remains a
post type, with the shared post detail and interaction routes.

Use "Altklausuren" consistently for the exam section: navigation, page heading,
browser title, breadcrumbs, search, counts and actions. Do not alternate between
"Archiv", "Archive", "Rekos" and "Klausurrekonstruktionen" as names for it.
Where a document's provenance matters, explain that it is a reconstruction, not
an official exam paper. A factual document-type description is not a second name
for the section. Do not imply that reconstructed documents are official papers.

Use "Fotos" consistently for the current photo section, with "Fotoalbum" for
an album. The permanent `/media` namespace and English `media` package stay
independent of the German display label. If video is introduced, choose an
equally explicit label such as "Fotos und Videos" without moving existing URLs.

Keep organizational membership, organizational functions and account permissions
distinct. A `MemberRole` is a function such as chair; it is not an account access
role. Use qualified names where a bare `Role` would be ambiguous.

## Actions and status

| URL action | Code term | German UI wording |
| --- | --- | --- |
| `create` | `Create` | Hinzufügen, or a specific action such as Altklausur einreichen |
| `edit` | `Edit` | Bearbeiten |
| `check` | `Check` | Angaben prüfen |
| `review` | `Review` | Einreichung prüfen |
| `vote` | `Vote` | Abstimmen |
| `invite` | `Invite` | Einladung senden |
| `login` | `Login` | Anmelden |
| `logout` | `Logout` | Abmelden |
| `register` | `Register` | Konto erstellen |

Contextual wording is appropriate: a publication check can say "Veröffentlichung
prüfen", and a submit button says "Einreichen" rather than "Prüfen". Name the
actual operation without introducing a competing domain term.

For submissions use `draft` / "Entwurf", `pending` / "In Prüfung", `approved` /
"Freigegeben" and `rejected` / "Abgelehnt". Approval actions say "Freigeben";
rejection actions say "Ablehnen". Preserve any distinct publication or account
states instead of forcing them into the submission status vocabulary.

## Code and file migration

Rename the owning package, imports, files, directories, exported types, local
variables, receiver names, methods, tests and fixtures together. Follow existing
Go casing and filename conventions. Use the canonical concept rather than a
literal transcription of every URL segment into a package.

The existing domain packages map as follows:

| Current package | Target package | Related naming changes |
| --- | --- | --- |
| `internal/model/archive` | `internal/model/exams` | `ArchivePage` → `ExamsPage`; `archive` variables → `exams` |
| `internal/model/forum` | `internal/model/posts` | `ForumPage` → `PostsPage`; `forum` variables → `posts` |
| `internal/model/gallery` | `internal/model/media` | `ImagesPage` → `MediaPage`; `gallery` variables → `media` |
| `internal/model/people` | `internal/model/members` | `Person` → `Member`; `Responsibility` → `MemberRole` where domain-equivalent |
| `internal/model/events` | unchanged | Keep event/calendar records distinct from albums |
| `internal/model/meetings` | unchanged | Keep meeting publications distinct from exam documents |
| `internal/model/organization` | unchanged | Organization identity and settings remain their own domain |
| `internal/model/users` | unchanged | User accounts and authentication remain their own domain |

The role functionality can remain in the members package. URL grouping does not
require a new package or a new service for every noun.

Examples of coordinated file renames:

- Controller/view `archive.go` → `exams.go`; domain submission handling stays
  `submissions.go`.
- Controller `forum.go` → `posts.go`; `gallery.go` → `media.go`.
- Templates `archive.html` → `exams.html`; `archivepreferences.html` →
  `examsettings.html`; `forum.html` → `posts.html`; `images.html` → `media.html`.
- Template `fachschaft.html` → `about.html`; matching page fields use `About`.
- Submission types and templates currently named `Request`, `RequestPage`,
  `requests.html`, `request.html` or `requestlist.html` get descriptive
  `Submission` names. Distinguish a submission list from an admin review queue.

Use `Submission` for the exam workflow, not `Request` or `Application`.
`http.Request` and ordinary network request variables remain requests. Likewise,
`Image` remains correct for image bytes, dimensions and processing; do not rename
every image variable to media. A media page and its image representation are
different concepts.

Review URL constants, template names and page fields, active-navigation keys,
CSS selectors, JavaScript hooks, anchor IDs, test names and accessible text as
part of the same change. Avoid leaving old terminology in hidden labels,
validation errors or email subjects. Use named types or context to distinguish
exam documents from organizational documents, rather than merging their models
because both URLs contain `documents`.

## UI acceptance criteria

Study-program names contain only the subject, for example `Informatik`, never
`Informatik (B. Sc.)`. `Abschluss` (`Bachelor` / `Master`) and `Prüfungsordnung`
are separate qualification dimensions. Migration 19 normalizes the stored
`programs.name` from `study_programs.name` while preserving legacy relationship
keys and public IDs. Multiple qualifications may share a program name.

Registration lists each `study_programs` record once using its public ID and
submits the degree separately as `bachelor` / `master`. The server resolves that
pair through `program_qualifications` to the existing account `program_id`.
Both optional fields may be omitted together; partial and unavailable pairs
produce field errors. Account management also displays program and degree
separately.

Registration optionally accepts a `version` from `program_versions` for the
selected study-program/degree pair. It initializes `user_archive_preferences`
in the account-creation transaction, using the qualification's public program
ID, English degree value and exact stored version. An omitted PO is never
guessed. `Abschluss` and `Prüfungsordnung` share a responsive two-column row.

Migration 20 stores PO years as four-digit SQLite integers in
`program_versions.name`, `archive.version` and `archive_assignments.version`.
Values such as `PO2023` become `2023`; no `PO` prefix is part of the stored
identity. Matching saved filter values are normalized in the same transaction.
HTML forms, query values and filter JSON use the corresponding decimal strings
(`"2023"`), and server validation checks that the chosen PO exists for the
selected qualification. Document edit/review versions are separate counters and
must not be changed by PO normalization. Historical source snapshots and original
file/storage identities remain untouched.

- Each section has the same canonical German name in the main navigation,
  sidenav, page heading, document title, breadcrumbs, home cards and sitemap.
- Links, forms, empty states, error/success messages, review screens and emails
  use the same nouns and status vocabulary.
- Singular/plural forms and grammatical inflections are allowed; competing
  section names and unexplained abbreviations are not.
- Update templates and controller-provided text together. A heading supplied by
  a handler must not disagree with a hard-coded heading inside its template.
- Keep original user content, legal names, document titles and provenance intact.
  These are records, not UI vocabulary to rewrite.

## Persistence and compatibility boundaries

Application naming changes must not rekey production records, rename stored
objects, alter original bytes or rewrite source provenance. Historic SQL
migrations, external formats and preserved originals are not targets for a
blind text replacement. Physical schema changes, if needed, require a new
forward migration; regenerate generated code from its owning source.

Known persisted terminology such as database `events` rows representing albums
needs an explicit adapter or migration plan, not a global replacement that
confuses calendar events and media albums. Canonical application models use the
new terms even where a storage adapter must refer to old schema names.

Storage helpers follow the canonical concepts too: `ExamSourceKey`,
`AlbumCoverSourceKey` and `AlbumCoverPreviewKey`. Their returned keys retain the
existing `archive/...` and `events/...` layouts; changing an API name must not
move or overwrite an object.

Preserve old fragment IDs when needed for incoming links, even when surrounding
UI and code names change. The proxy cannot translate browser-only fragments.
Compatibility names are acceptable only at documented storage/provenance and
proxy boundaries, not as parallel active vocabulary in the UI or domain API.

The migration review should inventory old terms across application code,
templates, assets, tests and documentation, and explain any retained occurrence.
Verify behavior after renaming; matching words alone does not prove semantic
equivalence.
