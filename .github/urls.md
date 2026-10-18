# URL convention

Status: canonical contract for the breaking URL and naming migration. Temporary
compatibility redirects belong in the front proxy; see the
[redirect plan](url-redirects.md).

The migration also includes packages, files, identifiers and UI wording. Follow
the [canonical naming contract](naming.md): code and URLs are English; UI labels
remain German and use one consistent term per concept.

## Principles

- Use clear English, lowercase path vocabulary and full query names. Use hyphens
  if a future descriptive segment needs multiple words.
- Use `create` and `edit` consistently. Use `check` for the submitter's final
  check and `review` for a reviewer's decision.
- Keep URLs shallow. Navigation groups do not dictate URL nesting.
- Use immutable 11-character public IDs for records, separate from internal
  database keys. Titles, storage keys and filenames do not determine a record's
  URL. Preserve existing internal identities and original document/image bytes.
- Use one canonical detail URL per record. News and discussion posts share
  `/posts/{post}`. Events and media albums are separate resources.
- Use no trailing slash except `/`. Static asset directory handling is separate.
- The application serves the canonical contract, without legacy route aliases.

The naming follows GOV.UK guidance on clear words, user tasks and unnecessary
words. English URLs with German UI are this project's policy; GOV.UK does not
require single-word segments. Opaque IDs and tokens retain their exact case.

## Identity and permanence

- Public IDs contain eight cryptographically random bytes, encoded with RFC 4648
  unpadded Base64url: exactly 11 case-sensitive ASCII characters. The final
  character is constrained by the canonical encoding; arbitrary 11-character
  strings are not necessarily valid IDs.
- Allocate once at persistent creation, including drafts. Enforce uniqueness per
  resource namespace, retry a generation collision, and retain the reservation
  after deletion. Never derive identity from clocks, titles, filenames or users.
- Integer database keys and foreign keys remain internal. New integer keys are
  database-generated. Existing database keys are preserved, not renumbered.
- Public IDs identify resources, not permission to access them. Ownership and
  capability checks remain mandatory on protected resources and actions.
- Sessions, invitations and account tokens are separate 256-bit secrets. Edit
  tokens identify an inspected version; neither uses public-resource entropy.
- Dates and ordering use explicit columns, never decoded or inferred from IDs.
- The section prefix is permanent too: changing a German label or menu position
  must not rename `/posts`, `/events` or other established namespaces.
- Informational pages remain source-controlled. Features declare routes and
  access policies in ordinary Go, and share route patterns and URL builders
  through `internal/urls`. No API namespace or configurable rewrite engine is
  part of this architecture.

## Tree

`{name}` is an immutable 11-character public record ID unless otherwise noted. Prefixes used only to
group routes do not need landing pages. An indented action does not imply that
its parent has a separate page: for example, roles need an editor, not an
additional read-only detail page.

```text
/                                      home
├── about
├── team
├── meetings                            published meetings
├── elections
├── regulations
├── contact
├── legal
├── privacy
├── accessibility
├── sitemap                             HTML site overview
│
├── community                           community overview
├── posts                               discussion listing
│   ├── create
│   └── {post}                          discussion or news post
│       ├── edit
│       ├── vote
│       └── comments                    add comment
│           └── {comment}
│               ├── edit
│               └── vote
├── news                                news listing
├── events
│   ├── create
│   └── {event}
│       ├── edit
│       └── cover                       image bytes
│
├── media                               album listing
│   ├── albums
│   │   └── {album}
│   │       └── cover/{size}            image bytes
│   └── {item}                          individual media page
│
├── exams                               past exams and reconstructions
│   ├── modules/{module}
│   ├── documents/{document}            protected PDF
│   ├── settings                       saved filters
│   └── submissions                    own submissions
│       ├── create
│       └── {submission}                details and status
│           ├── edit
│           └── check                   check and submit
│
├── users/{user}                        public profile
├── login
├── logout
├── register
├── password
│   ├── forgot
│   └── reset
├── email/confirm
├── invitation/accept
│
├── admin
│   ├── settings
│   │   ├── identity
│   │   ├── contact
│   │   ├── legal
│   │   └── links
│   ├── members
│   │   ├── create
│   │   └── {member}
│   │       ├── edit
│   │       └── invite
│   ├── roles
│   │   ├── create
│   │   └── {role}/edit
│   ├── meetings
│   │   ├── create
│   │   └── {meeting}
│   │       ├── edit
│   │       └── check                   check and publish
│   ├── submissions                    review queue
│   │   └── {submission}
│   │       ├── edit
│   │       └── review                  approve or reject
│   └── users
│       └── {user}
│           ├── reset                  request password reset email
│           └── confirm                send confirmation email
│
├── documents/{document}                organizational PDF
├── images/{image}/{size}                image bytes
├── assets/{path...}                    deployed static assets
└── healthz
```

`/community` links to `/posts`, `/news`, `/events` and `/media`; these sections
remain independently addressable at the root. The discussion navigation label
can be "Forum", but its canonical path is `/posts`, not `/forum` or `/d`.

`/media/{item}` initially displays photos. The name allows later video support
without changing item URLs; video delivery is outside this URL change.

Image variants retain the current allow-list: `original`, `200`, `400`, `600`,
`800`, `1200` and `1600`. Album covers allow the numeric variants only.
`/events/{event}/cover` serves the event's uploaded cover, independently of an
album's cover.

`/sitemap` remains HTML. An XML sitemap and robots file are separate future
features, not implied routes in this change.

## Methods and journeys

Read-only pages and files use GET and HEAD. Ordinary forms use GET to display
and POST to process. Existing-record editors use `/edit`; singleton settings
forms use their own section URL directly.

| Route | GET | POST |
| --- | --- | --- |
| `.../create` | Creation form | Create record |
| `.../edit` | Editing form | Save changes |
| `/exams/submissions/{submission}/check` | Check answers | Submit |
| `/admin/meetings/{meeting}/check` | Check publication | Publish |
| `/admin/submissions/{submission}/review` | Review submission | Approve or reject |
| `/password/forgot` | Request form | Send reset email |
| `/password/reset` | Token/password form | Reset password |
| `/email/confirm` | Token confirmation form | Confirm email |
| `/invitation/accept` | Invitation form | Accept invitation |
| `/admin/users/{user}/reset` | Confirm destination | Send reset email |

Voting, comment creation/editing, member invitation sending, admin confirmation
sending and logout use POST only. Comment editing stays inline on the post page.
An admin review uses a form field to select approval or rejection, rather than
separate action endpoints.

Successful form submissions redirect with 303 to the relevant record or list.
The submission detail page supplies the receipt, reference and next steps after
submission, and remains useful when bookmarked. GET never submits, publishes or
consumes an account token.

All protected routes have explicit authentication, authorization, CSRF and body
limit policies. Policies must not depend on old path-prefix string checks.

## Queries

Use English words for names and application-defined values. IDs, opaque tokens
and user-entered text retain their exact values and case. These are the shared
names; a route accepts only parameters it actually supports.

| Name | Meaning | Example |
| --- | --- | --- |
| `search` | Search text | `/posts?search=welcome` |
| `page` | Page number starting at 1 | `/media?page=2` |
| `program` | Study program public ID, or explicit `all` | `/exams?program=m7Qp2Vx9aK4` |
| `degree` | Degree filter | `/exams?degree=bachelor` |
| `version` | Program version | `/exams?version=2025` |
| `year` | Calendar year | `/meetings?year=2026` |
| `type` | Creation type | `/posts/create?type=news` |
| `status` | Submission filter, when supported | `/admin/submissions?status=pending` |
| `sort` | Sort order, when supported | `/events?sort=oldest` |
| `revision` | Existing exam document public ID to revise | `/exams/submissions/create?revision=8vJd3Rk5nQ0` |
| `token` | Account token | `/email/confirm?token=...` |
| `date` | Meeting creation date | `/admin/meetings/create?date=2026-10-18` |
| `number` | Meeting creation number | `/admin/meetings/create?number=12` |

- Omit empty values and defaults such as `page=1` from generated links.
- Repeat keys for multiple selections:
  `program=m7Qp2Vx9aK4&program=8vJd3Rk5nQ0`.
- Keep ordering deterministic, including a stable record-ID tie-breaker.
- Use explicit English values, such as `draft`, `pending`, `approved`, `rejected`,
  `news` and `discussion`. Map storage values if necessary; do not change
  production data merely to rename URLs.
- Do not emit `q`, `seite`, `jahr`, `datum`, `nummer`, `applied`, `scope`,
  `choose_scope` or `edit_token` in new URLs. Keep edit/concurrency tokens in
  server state or form fields, not navigation links.
- Saved exam preferences select initial filters. Generated results links must
  include the effective filters. Explicit filter submissions, including clearing
  all filters, must override preferences without an `applied` or `scope` flag.
  Use `program=all` when no qualification filters are selected. This meaningful
  filter value survives module/pagination links and prevents saved defaults from
  being silently reapplied; it is not omitted as a default parameter.
- Success messages belong in session flash state or the record's actual status,
  not query flags such as `gespeichert=1`.
- Review forms carry the inspected submission version in a hidden form field.
  A decision submitted after the content or review state changes returns 409;
  reviewers must inspect the current version before deciding again.
- Query parameters select content or a form; mutations use POST form fields.
- Preserve existing fragment identifiers for comments, images and team sections
  where practical. Fragments are browser-local and cannot be read by the proxy.

Examples:

```text
/exams?search=mathematics&program=m7Qp2Vx9aK4&degree=bachelor&page=2
/posts/create?type=news
/exams/submissions/create?revision=8vJd3Rk5nQ0
/meetings?year=2026
```

## Implementation scope

Update route registration, explicit route policies, handler dispatch, URL
builders, templates, form actions, breadcrumbs, pagination and email links
together. Audit database-backed internal links without rewriting original source
URLs or provenance. Historical documents use stored random public IDs keyed by
their preserved original filenames. Export the registry-backed resource and
filename mappings for the proxy before cutover.

`internal/publicid` owns generation, canonical validation, allocation, lookup and
resolution. Migrations 17 and 18 add permanent reservations and database-backed
integer sequences without changing original records. New integer resources call
`NextInteger`, insert the returned key, then `Ensure`, all in one transaction.
Read paths use `Lookup` or `Resolve`; they do not invent IDs while rendering.
Future resource types add their table and registry/sequence registration in a
new forward migration; never extend a historical migration's table inventory.

## Canonicalization and pagination

- Recognized GET/HEAD routes normalize trailing slashes and default `page=1`
  through permanent 308 redirects. Application migration aliases stay at the
  proxy, with its separate rollout policy.
- Noncanonical public IDs, including old decimal IDs, return 404. Do not guess
  targets or change the case of an ID. Unknown paths return 404.
- Empty, malformed, negative, zero, duplicate and overflowing page values return
  400. A valid positive page beyond the collection's range returns 404. An empty
  collection has a valid first page.
- Indexable public HTML has a root-relative canonical link generated from its
  declared query contract, never an untrusted request Host. Effective exam
  filters, meeting years and page numbers greater than one retain distinct
  addresses. A post's search/back-navigation context is not its identity.
- Errors, HTML fragments, management pages, private workflows and account forms
  have no canonical link. Canonical metadata is not an access-control mechanism.

Verify the route contract, methods, permissions, CSRF checks, token journeys,
query parsing and generated links. Legacy paths should return 404 from the new
application; only the front proxy provides the temporary mappings.

## References

- [GOV.UK URL standards](https://guidance.publishing.service.gov.uk/writing-to-gov-uk-standards/plan-manage-content/follow-url-standards/)
- [GOV.UK service naming](https://www.gov.uk/service-manual/design/naming-your-service)
- [Check answers](https://design-system.service.gov.uk/patterns/check-answers/)
- [Confirmation pages](https://design-system.service.gov.uk/patterns/confirmation-pages/)
