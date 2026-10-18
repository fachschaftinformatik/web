# URL durability and extension review

Reviewed: 18 October 2026. Scope: the current migration working tree, GOV.UK
guidance and WordPress developer documentation/source. The findings below record
the pre-extension audit; the adopted design is the [canonical contract](urls.md).
The subsequent implementation separates preserved database keys from immutable
11-character random public IDs and centralizes route/link definitions.

## Conclusion

Keep the main URL tree. Its stable record IDs, shallow resource paths, separate
collections and explicit workflow actions are a good foundation. Future growth
should add routes without changing existing addresses.

The implementation now uses shared route/link definitions, canonicalization,
stable public query identities, feature-owned namespaces and extension tests.
A configurable permalink engine would introduce more moving parts than this
application needs. Production proxy cutover remains a deployment task.

## What GOV.UK actually recommends

The current [URL standards][gds-urls] call for lowercase, clear words, short and
unambiguous addresses, verb stems, and hyphens between words. They recommend
user needs over changeable organisation, scheme or service names. The Service
Manual also says URLs should help users understand where they are [gds-ui].

Applied here:

- `/events`, `/media` and `/exams` identify durable concepts.
- `/posts/{id}` survives a title edit or reclassification from discussion to news.
- `/community` can remain a navigation landing page without owning every child
  section's URL. Moving a navigation item need not move its content.
- `create`, `edit`, `check` and `review` make distinct tasks explicit. GOV.UK does
  not prescribe these exact route suffixes or a universal CRUD tree.
- Single-word segments are our preference, not a GOV.UK requirement. A future
  multiword concept should use an intelligible hyphenated phrase if approved,
  rather than an acronym, invented compound or misleading short word.
- English URLs with German UI labels are our bilingual convention. GOV.UK's
  English vocabulary guidance is not evidence that English URLs are inherently
  more usable for a German audience.
- Opaque IDs trade some readability for permanence. Retain that deliberate
  choice for records; readable static page names remain appropriate.

No recommendation requires putting menu hierarchy, publication dates, authors,
categories or implementation package names into record URLs.

## How WordPress routes content

WordPress has two related routing systems, not one universal URL convention.

### Public permalinks

1. A post has an internal ID and a stored slug. WordPress supports both query
   forms such as `?p=123` and configured permalink structures using IDs, slugs,
   dates, categories and authors [wp-permalinks].
2. The Rewrite API maps public paths to WordPress query variables. Plugins can
   add rules, endpoints and permalink structures. Rewrite rules are regenerated
   when configuration changes, not on every request [wp-rewrite].
3. `register_post_type` separates the internal type key, translated labels,
   public rewrite base, archive availability, capabilities and REST exposure
   [wp-types]. Here, "archive" means a collection listing, not our exam section.
4. `get_permalink` is the common outward-link interface; callers do not all
   reproduce the configured permalink structure [wp-links].
5. `redirect_canonical` normalizes many alternative URLs on GET/HEAD. It also
   contains 404-guessing logic. `wp_old_slug_redirect` can resolve recorded old
   slugs for supported non-hierarchical types; it is not a complete redirect
   history for arbitrary moves, page hierarchies or plugin changes
   [wp-canonical] [wp-old-slugs].

The strongest transferable ideas are identity independent of presentation,
central outward-link generation, explicit content registration and deliberate
canonicalization. Date/category-dependent permalinks, arbitrary plugin rewrite
ordering and guessed redirects would make this site's permanence harder.

### Server-rendered extension contract

This application has no planned API. Features register their HTML pages and form
handlers directly, with explicit methods, validation and access policies.
Public pages remain unversioned; forms use GET/POST and 303 redirects. Namespace
ownership prevents features from shadowing one another. Routing remains distinct
from permission checks on the actual record.

## Findings before the extension implementation

### 1. Settle query identities and values before release — high priority

The contract documents `program` as an ID and `degree=bachelor`. The current
catalogue emits program display names and capitalized degree labels, then
compares them as exact strings. See `site/exams.go`,
`model/exams/pagination.go` and `view/exams.go`.

Observed on the isolated production-derived development copy:

| Request | Result |
| --- | --- |
| `/exams?degree=bachelor` | 200, no modules |
| `/exams?degree=Bachelor` | 200, 33 modules |
| `/exams?program=12` | 200, no modules |

Recommendation: use stable qualification IDs or immutable public codes for
`program`, and English enum values such as `bachelor` and `master` for `degree`.
Map these to German labels and existing persistence values at the boundary.
Study-program labels must be editable without invalidating saved URLs. Define
whether `version` identifies an actual regulation record or a cross-program
year label; those are different filters. Preserve existing stored settings with
an adapter or forward migration when changing their external representation.

Keep `program=all` as the explicit override of saved qualification defaults.
Specify that it is exclusive with specific program IDs; conflicting input must
not silently become an accidental combination. Query values that are sets
should be deduplicated and sorted when generating links.

### 2. Complete canonical URL handling — high priority

`site/routes.go` explicitly registers canonical route shapes,
but the overall canonicalization policy is incomplete:

- A real `/posts/{id}` and `/posts/0{id}` both returned 200 in the review.
- `/posts?page=1`, `/posts?page=999999` and `/media?page=999999` returned 200.
  Model-level clamping can make many addresses display the same page.
- `/about/` returns 404. This follows the current strict spelling policy, but
  there is no agreed friendly-normalization boundary for canonical paths.
- `view/templates/layout.html` has no canonical-link metadata.

Recommendation: define one canonical public origin and one address for each
public representation. Require canonical decimal IDs on numeric-ID routes and
exact case for opaque IDs. Never lowercase tokens or user text globally.

For known canonical routes, normalize harmless GET/HEAD spelling differences
with one-hop permanent redirects where the target is unambiguous. For malformed
or out-of-range pagination, choose and test one policy; 404 for nonexistent page
numbers is a simple consistent option. Do not make clamping create unlimited
200 aliases. POSTs must not be replayed through normalization redirects.

Add canonical metadata for indexable public HTML, generated from the trusted
origin and typed page state. Preserve meaningful page/filter state rather than
pointing every pagination or search result at page 1. Decide separately which
search/filter combinations should be indexed. Private, admin and token pages
need an explicit indexing policy; canonical tags are not access control.

### 3. Share route definitions and outward-link construction — high priority

Route patterns live in `site/routes.go`; record URL methods live
in `view/records.go`; controllers, email generation and templates also
contain path literals. Existing builders help, but there is no shared route
identity joining registration and link generation.

Recommendation: introduce a small dependency-light Go package, for example
`internal/urls`, with named route definitions and typed builders. It must be
usable by controllers, views and email generation without importing view code.
Examples of outward interfaces:

```go
urls.Post(postID)
urls.EditPost(postID)
urls.MediaItem(itemID)
urls.ExamModule(moduleID, filters)
urls.CreateSubmission(revisionID)
```

Registration should consume the same declared patterns. Validate/escape each
path parameter and encode queries with `net/url`. Do not accept arbitrary path
fragments or arbitrary action suffixes as the interface. An independent contract
test must still freeze the intended external addresses; sharing a wrong constant
between router and builder is not sufficient verification.

### 4. Make adding a feature an additive operation — medium priority

Adding a section currently involves route registration, the generic admin
dispatcher in `site/users.go`, renderer registration in
`view/render.go`, page fields, navigation, sitemap and breadcrumbs.
`Page` also aggregates every domain's page state.

Recommendation: use ordinary feature registration functions, grouped by domain,
and a small public-section catalogue for shared navigation/sitemap metadata.
Each feature registers concrete handlers and an explicit policy. Feature-specific
breadcrumbs and workflow logic remain feature-specific. Keep this in Go; a
runtime plugin loader or database-configured regex router is unnecessary.

Require an explicit public/protected policy, methods, parameter rules and query
contract for every new route. A missing policy should fail registration or a
contract check, rather than silently becoming public through a zero value.
Keep model-level ownership and permission checks as well.

### 5. Define namespace ownership and ID grammars — medium priority

Reserve existing root sections and infrastructure space such as `/api`,
`/assets`, `/images`, `/documents`, `/.well-known`, `/robots.txt` and
`/sitemap.xml`. Reserving names does not mean implementing those endpoints now.

Within a collection, distinguish action words from valid record identifiers.
The fixed 11-character canonical Base64url grammar makes `/posts/create` and a
future `/posts/feed` unambiguous. New actions must not be valid public IDs.

`/media/{item}` and `/media/albums/{album}` can coexist safely if item IDs have
an explicit grammar. Future video pages can retain `/media/{item}`; delivery
routes and storage adapters can be added separately. The current implementation
still serves photos, so video support will require domain/rendering work.

`/admin/submissions` currently belongs to exams. Preserve that ownership unless
a deliberate cross-domain queue is designed. A new media-submission workflow
can use `/admin/media/submissions` without moving the exam queue. Likewise,
`/admin/roles` means member functions, not account permissions.

Informational pages remain source-controlled. A future section such as `/guides`
can register explicit Go handlers and templates. Menu nesting must not
automatically become a page's permanent path.

### 6. Reconcile temporary redirects with long-term bookmarks — high priority

The current [proxy plan](url-redirects.md) deliberately uses temporary 307
redirects with `no-store` and a removal phase. That is useful during a reversible
cutover, but removing a mapping breaks the corresponding old bookmark.

Recommendation: distinguish reversible rollout mappings from enduring link
preservation. After targets are verified, consider retaining exact GET/HEAD
mappings long-term as permanent redirects at the proxy. This would amend the
current temporary-only policy and needs an explicit decision. Do not replay old
mutations or guess replacement records.

The old photo-gallery `/events` namespace still conflicts with the new calendar
namespace. No routing framework can disambiguate identical old/new URLs on one
host. A legacy hostname or explicit acceptance of broken ambiguous bookmarks
remains necessary. Keep filename/event-editor exports as deployment artifacts.

## Adopted extension acceptance contract

An extension is ready when:

1. It claims a nonconflicting namespace and has named routes and typed builders.
2. Existing URL contract cases pass unchanged; published public IDs still resolve.
3. Changing titles, UI labels, menu placement or taxonomy does not change detail
   URLs. Adding another album relationship must not move an item's detail URL.
4. Every route has explicit methods, policy, ID grammar, query semantics and
   indexability. Unknown record IDs and unknown paths have defined responses.
5. Builder output resolves to the intended handler. Contract cases cover reserved
   words, ID aliases, slash spelling, repeated filters and pagination boundaries.
6. Public links, emails, breadcrumbs and sitemap entries use the same builders.
7. Existing GET/POST, CSRF, ownership, token and concurrency journeys remain valid.
8. Any redirect has an exact target, preserves required values, has no loop or
   chain, and is tested at the proxy/application boundary.

Useful extension exercises: add a source-controlled guide, add a video item,
or add an event category. None should require renaming the existing public tree
or copying routing logic into several unrelated layers.

## Recommended sequence

1. Resolve the filter-value mismatch and choose canonicalization/indexing rules.
2. Add the shared route/builders package and independent URL contract cases.
3. Group feature registrations and share public-section metadata.
4. Recheck proxy retention and the `/events` collision before production cutover.

Keep the chosen record paths and immutable public-ID format. The agreed design
does not introduce editable page models, title slugs or permalink settings.

## Sources

Primary guidance and reference implementations inspected for this review:

- [GOV.UK URL standards][gds-urls]
- [GOV.UK writing for interfaces: URLs][gds-ui]
- [WordPress permalink structures][wp-permalinks]
- [WordPress Rewrite API][wp-rewrite]
- [WordPress content type registration][wp-types]
- [WordPress permalink generation][wp-links]
- [WordPress canonical redirect implementation][wp-canonical]
- [WordPress old-slug redirect implementation][wp-old-slugs]
- [WordPress REST routes, namespaces and endpoints][wp-rest]

[gds-urls]: https://guidance.publishing.service.gov.uk/writing-to-gov-uk-standards/plan-manage-content/follow-url-standards/
[gds-ui]: https://www.gov.uk/service-manual/design/writing-for-user-interfaces#urls
[wp-permalinks]: https://wordpress.org/documentation/article/customize-permalinks/
[wp-rewrite]: https://developer.wordpress.org/apis/rewrite/
[wp-types]: https://developer.wordpress.org/reference/functions/register_post_type/
[wp-links]: https://developer.wordpress.org/reference/functions/get_permalink/
[wp-canonical]: https://developer.wordpress.org/reference/functions/redirect_canonical/
[wp-old-slugs]: https://developer.wordpress.org/reference/functions/wp_old_slug_redirect/
[wp-rest]: https://developer.wordpress.org/rest-api/extending-the-rest-api/routes-and-endpoints/
