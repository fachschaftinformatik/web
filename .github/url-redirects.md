# Temporary proxy redirects

This is the redirect inventory for the [target URL convention](urls.md), reviewed
against the current route registrations, controller dispatch, URL builders and
templates on 2026-10-18. It is a deployment plan, not active proxy configuration.
The application migration and proxy rollout are separate work.

The proxy owns compatibility. The new application exposes canonical routes only.
The inventory covers routes supported by the current application, including its
legacy aliases. Older externally published URLs not present in these sources
need an additional inventory from production access logs and stored links.

## Redirect policy

- Match GET and HEAD only. Use temporary **307** redirects during the migration,
  with `Cache-Control: no-store` so mappings can be corrected or removed.
- Do not redirect or replay legacy POST, PUT, PATCH or DELETE requests. Reject
  retired mutation URLs with 404; users must reload the canonical form. Preserve
  authentication and CSRF checks at the destination.
- Send each old URL directly to its final destination, combining path and query
  changes in one hop. Do not route through existing application redirects.
- Translate queries as described below, including on unchanged paths. Preserve
  repeated values and encode them once. Preserve tokens exactly.
- Prefer explicit paths and constrained captures over blanket prefix rewrites.
  Match literals such as `new`, `files`, `requests` and `neu` before ID captures.
- Do not redirect unknown paths to the home page. Unknown IDs remain 404 at the
  destination. Avoid rules whose source also matches their canonical destination.
- Canonical routes win over ambiguous legacy meanings. In particular, see the
  `/events` collision below before deploying any rules.

The tables use `{id}`-style placeholders, not proxy regex syntax. A record capture
must be translated through the exported `public_ids` registry, not copied into
the new URL. Internal database IDs remain unchanged; public record IDs are now
11-character random values. Export the relevant namespace's `record_id` and
`public_id` after migrations 17 and 18, and keep that database and its reservations
with all backups. Token values and semantic values such as years stay exact.

For example, an old `/users/42` maps to `/users/{public ID for users/42}`. The
old numeric address is not an application alias. Namespace mappings also apply
to module, revision and program IDs in queries. Never regenerate public IDs
independently at the proxy.

## Public pages and account links

| Old GET/HEAD path | Destination |
| --- | --- |
| `/fachschaft` | `/about` |
| `/fachschaft/sitzungen` | `/meetings` |
| `/fachschaft/wahlen` | `/elections` |
| `/fachschaft/ordnungen` | `/regulations` |
| `/impressum` | `/legal` |
| `/datenschutz` | `/privacy` |
| `/barrierefreiheit` | `/accessibility` |
| `/forgot` | `/password/forgot` |
| `/reset` | `/password/reset` |
| `/verify` | `/email/confirm` |
| `/invite` | `/invitation/accept` |

Preserve `token` on reset, confirmation and invitation links, including links in
already-sent emails. The target GET only displays a form; account token use still
requires POST. Do not log token values in redirect verification records.

`/`, `/team`, `/contact`, `/sitemap`, `/community`, `/users/{id}`, `/login`,
`/register`, `/documents/{id}`, `/healthz` and ordinary `/assets/...` paths retain
their meanings. `/logout` remains POST only and needs no redirect.

## Posts, news and calendar events

| Old GET/HEAD path | Destination |
| --- | --- |
| `/discussions` | `/posts` |
| `/discussions?type=news` | `/news` (remove `type`) |
| `/discussions?type=event` | `/events` (remove `type`) |
| `/d/new` | `/posts/create` |
| `/d/{post}` | `/posts/{post}` |
| `/d/{post}/edit` | `/posts/{post}/edit` |
| `/community/neuigkeiten` | `/news` |
| `/community/neuigkeiten/new` | `/posts/create?type=news` |
| `/community/veranstaltungen` | `/events` |
| `/community/veranstaltungen/new` | `/events/create` |
| `/community/veranstaltungen/{event}` | `/events/{event}` |
| `/community/veranstaltungen/{event}/edit` | `/events/{event}/edit` |
| `/community/veranstaltungen/{event}/cover` | `/events/{event}/cover` |

The `type` conditions precede the ordinary discussion-list rule. Translate
remaining supported query parameters in the same redirect. A fixed destination
such as `type=news` cannot be overridden by an incoming parameter.

`/d` itself is not currently registered, but an exact GET/HEAD convenience rule
may send it to `/posts`. `/forum` was discussed as a label, not chosen as a path;
do not introduce an additional alias without evidence of published links.

Older event posts still have post IDs and comment threads. Keep `/d/{post}`
mapped to `/posts/{post}` by default, preserving those threads. Calendar events
have a separate `legacy_post_id` relationship; do not assume a post ID equals a
calendar event ID. Resolve old event-post editors to `/events/{event}/edit` with
an exported post-to-event map where the current editor delegates to a calendar
event. Ordinary post editors use the structural rule above.

Old vote/comment endpoints are POST only and get no redirect. Update every form
action in the application migration.

## Exams

Apply special cases before the generic module and obsolete-entry patterns.

| Old GET/HEAD path | Destination |
| --- | --- |
| `/archive` | `/exams` |
| `/archive/new` | `/exams/submissions/create` |
| `/archive/preferences` | `/exams/settings` |
| `/archive/files/{file}` | `/exams/documents/{file}` |
| `/api/v1/archive/{file}/file` | `/exams/documents/{file}` |
| `/archive/files/{file}/revision` | `/exams/submissions/create?revision={file}` |
| `/archive/requests` | `/exams/submissions` |
| `/archive/requests/{submission}` | `/exams/submissions/{submission}` |
| `/archive/requests/{submission}/edit` | `/exams/submissions/{submission}/edit` |
| `/archive/requests/{submission}/review` | `/exams/submissions/{submission}/check` |
| `/archive/{module}` | `/exams/modules/{module}` |
| `/archive/{module}/{entry}` | `/exams/modules/{module}` |

The last rule matches the current obsolete-entry behavior: it opens the module,
not an invented document mapping. Restrict module captures to valid ID syntax or
an exported module list, so unknown special paths do not become module redirects.
Keep PDF authentication, permission checks and private caching at the target.

## Media pages and bytes

| Old GET/HEAD path | Destination |
| --- | --- |
| `/community/events` | `/media` |
| `/community/events/{album}` | `/media/albums/{album}` |
| `/community/events/{album}/{image}` | `/media/{image}` |
| `/media/{image}/{size}` | `/images/{image}/{size}` |
| `/media/events/{album}/cover/{size}` | `/media/albums/{album}/cover/{size}` |

For the old image-byte rule, constrain `{image}` to canonical positive decimal
IDs and `{size}` to `original`, `200`, `400`, `600`, `800`, `1200`, `1600`.
For album covers, allow only numeric sizes from that list. A generic
`/media/{anything}/{anything}` rule would intercept new album URLs.

The old album/image page identifies an image within an album; export valid pairs
if enforcing the old membership check at the proxy. Do not assume an album ID is
an image ID. Existing image IDs remain the media item IDs.

### Collision: `/events` used to mean photo albums

The current `/events` and `/events/{rest...}` routes redirect to
`/community/events...`. The new `/events` is the calendar event listing.

These old and new requests are indistinguishable on the same host:

| Old meaning | Desired old-link destination | New meaning |
| --- | --- | --- |
| `/events` | `/media` | Calendar event listing |
| `/events/{album}` | `/media/albums/{album}` | Calendar event detail |

**Do not install these redirects on the canonical host.** It is impossible to
preserve both meanings for identical URLs. The new calendar routes take priority;
some old gallery bookmarks will therefore show calendar content or return 404.

If an old hostname can be dedicated to legacy traffic, host-scoped rules can map
its `/events` to `/media` and `/events/{album}` to `/media/albums/{album}`.
Without that separation, only exact, verified paths that cannot collide with new
routes can be supported. A legacy three-segment image page
`/events/{album}/{image}` can map to `/media/{image}` when both IDs are constrained
to valid decimal IDs; exclude new action words such as `edit` and `cover`.
Never infer request age from cookies or referrers.

## Administration

The destination column describes the correct landing page for old GET/HEAD
links. Old POST actions are not forwarded.

| Old GET/HEAD path | Destination |
| --- | --- |
| `/verwaltung` | `/admin` |
| `/verwaltung/stammdaten` | `/admin/settings` |
| `/verwaltung/stammdaten/identitaet` | `/admin/settings/identity` |
| `/verwaltung/stammdaten/kontakt` | `/admin/settings/contact` |
| `/verwaltung/stammdaten/rechtliches` | `/admin/settings/legal` |
| `/verwaltung/stammdaten/verknuepfungen` | `/admin/settings/links` |
| `/verwaltung/mitglieder` | `/admin/members` |
| `/verwaltung/mitglieder/neu` | `/admin/members/create` |
| `/verwaltung/mitglieder/{member}` | `/admin/members/{member}/edit` |
| `/verwaltung/funktionen` | `/admin/roles` |
| `/verwaltung/funktionen/neu` | `/admin/roles/create` |
| `/verwaltung/funktionen/{role}` | `/admin/roles/{role}/edit` |
| `/verwaltung/referate`, `/verwaltung/projektstellen`, `/verwaltung/aufgaben` | `/admin/roles` |
| `/verwaltung/{kind}/neu` for those three legacy kinds | `/admin/roles/create` |
| `/verwaltung/{kind}/{role}` for those three legacy kinds | `/admin/roles/{role}/edit` |
| `/verwaltung/sitzungen` | `/admin/meetings` |
| `/verwaltung/sitzungen/neu` | `/admin/meetings/create` |
| `/verwaltung/sitzungen/{meeting}` | `/admin/meetings/{meeting}/edit` |
| `/verwaltung/sitzungen/{meeting}/pruefen` | `/admin/meetings/{meeting}/check` |
| `/verwaltung/einreichungen` | `/admin/submissions` |
| `/verwaltung/einreichungen/{submission}` | `/admin/submissions/{submission}/review` |
| `/verwaltung/einreichungen/{submission}/edit` | `/admin/submissions/{submission}/edit` |
| `/verwaltung/antraege` | `/admin/submissions` |
| `/verwaltung/antraege/{submission}` | `/admin/submissions/{submission}/review` |
| `/verwaltung/benutzer` | `/admin/users` |
| `/verwaltung/benutzer/{user}` | `/admin/users/{user}` |
| `/verwaltung/benutzer/{user}/passwort-zuruecksetzen` | `/admin/users/{user}/reset` |

`/verwaltung/mitglieder/{member}/einladen` and
`/verwaltung/benutzer/{user}/bestaetigung-senden` are POST only. Their replacements
are `/admin/members/{member}/invite` and `/admin/users/{user}/confirm`, not proxy
redirects. Preserve access controls regardless of the new English section names.

Legacy responsibility kinds selected a default group when creating a role.
Forward to the generic role creation form and require an explicit group choice;
do not silently convert all legacy groups into departments.

## Historical document filenames

`/assets/content/{file}` serves a PDF from `historical_meeting_documents` by its
`file` column. It is not a static asset. New publications already use
`/documents/{id}` with a different identifier space.

Historical records receive random public IDs in registry namespace
`historical_documents`, keyed by their exact original database filename. Newer
publications use namespace `publications`, keyed by their preserved internal ID.
Both share `/documents/{id}`, with a cross-namespace uniqueness constraint for
that URL space. Export an exact map from the migrated production catalogue:

```text
/assets/content/{old filename} -> /documents/{historical document ID}
```

Join the exact database filename to the registry, without normalizing case or
decoding it twice. Do not strip `.pdf` and assume the result is a publication ID. Match
only inventoried filenames; unknown filenames remain 404. Preserve bytes,
checksums, original filenames, storage keys and source provenance. Map old
publication-ID `/documents/{id}` links to their new registry-backed public IDs.

## Query translations

Translate parsed query parameters, not substrings of the raw URL. Re-encode
names and values once. Apply changes alongside the path redirect, or on an
unchanged canonical path only when a legacy parameter is actually present.

| Old name/value | Canonical treatment |
| --- | --- |
| `q` | Rename to `search` |
| `seite` | Rename to `page` |
| `jahr` | Rename to `year` |
| `datum` | Rename to `date` on meeting creation forms |
| `nummer` | Rename to `number` on meeting creation forms |
| `page`, `version`, `token` | Retain where supported |
| `program` | Translate legacy program IDs/display names to public program IDs; preserve `all` |
| `degree=Bachelor` / `degree=Master` | Translate to `bachelor` / `master` |
| `type=news` on `/discussions` | Select `/news`, remove `type` |
| `type=event` on `/discussions` | Select `/events`, remove `type` |
| `type=discussion` on list/detail links | Remove redundant type |
| `applied=1` with no qualification filters on exam result pages | Replace with `program=all` |
| Other `applied`, `scope`, `choose_scope` | Remove UI control flags |
| `edit_token` | Remove from URLs; destination must obtain a fresh edit token |
| `gespeichert`, `eingeladen`, `einladung`, `zuruecksetzen`, `bestaetigung`, `veroeffentlicht` | Remove success-message flags |

Pass only supported destination parameters. Keep repeated filters. For a
single-valued parameter supplied in both old and new forms, prefer the canonical
name, with at most one resulting value. A fixed route mapping controls `type`
and `revision`; remove competing input values rather than appending duplicates.
Do not lowercase or otherwise alter search text, IDs or tokens.

Preserve the old intent to clear saved preferences by translating `applied=1`
without qualification filters into `program=all`, not a bare `/exams` URL that
would restore saved defaults. The same explicit filter value remains on result,
module and pagination links. Dropping `choose_scope` reopens the normal editor
rather than reproducing its old UI step.

Fragments such as `#comment-42`, `#image-42` and `#team-board` never reach the
proxy. Preserve existing HTML anchor IDs where practical; fragment changes
cannot be implemented as general server-side rewrite rules.

## Cutover and removal

1. Implement and verify canonical application routes, policies, queries and
   emitted links. Apply the registry/sequence migrations on the deployment copy.
2. Export public-ID maps for all exposed resources, exact historical filenames,
   and delegated event-editor relationships. Review
   access logs, stored internal links and outstanding email links for additional
   historical paths and hostnames. Keep private data and actual maps outside
   tracked documentation.
3. Resolve `/events` ownership as above. Install constrained, ordered proxy rules
   alongside the new application deployment, without first sending users to
   routes that do not exist.
4. Check GET and HEAD, query translation, repeated filters, token integrity,
   file Content-Type and Range behavior, authorization at the destination,
   single-hop Location headers, absence of loops and rejection of old POSTs.
   Confirm new `/events`, `/media/albums/{id}` and `/media/{id}` bypass legacy
   rewrites. Verify unknown paths remain 404.
5. Record the activation date and review usage after 30 days. Set a removal date
   based on observed traffic and remaining token lifetimes; no fixed retirement
   date is assumed by this document. Remove unused proxy rules without changing
   the canonical application contract.
