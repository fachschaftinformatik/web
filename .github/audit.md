# Audit logging

`internal/audit.Audit(ctx, event, outcome, userID, details...)` is the shared
structured logging function. Events go through `log/slog` to the application's
log output and are appended to the SQLite `audit_events` table for requests and
background maintenance.

```go
audit.Audit(ctx, "submission.reviewed", "success", reviewerID,
    slog.String("submission_id", submissionID),
    slog.String("status", "approved"))
```

Emit successful state changes after their database transaction commits. Use
fixed event and outcome names, account/resource identifiers, and bounded metadata.
Never pass passwords, password hashes, email local parts, cookies, CSRF values,
recovery/verification links, request bodies, or raw query strings as details.

## Request correlation

Registered application routes assign a fresh server-generated `request_id`,
returned in `X-Request-ID`. Request-scoped audit events include the HTTP method
and registered route pattern. Authenticated requests also include
`actor_user_id`; `user_id` identifies the account concerned by an event. For
example, an administrator requesting another user's password reset has distinct
actor and subject identifiers.

The database retains the UTC timestamp, event/outcome, actor and subject account
IDs, request ID, method, route, and selected JSON details. Account identifiers
have no foreign keys, so deleting an account preserves its history. The audit
store only inserts records.

Business transactions complete before audit writes. A failed audit write is
reported as `audit persistence failed` in the application log without rolling
back a completed action or breaking login. Persistence has a separate bounded
timeout and also records events when the original request has been canceled.

Every POST and every application response with status 400 or higher records an
`http.request` outcome and status. Known resource path parameters identify the
affected resource. This covers forum posts, comments, votes, events, gallery
changes, settings and administrative actions. Successful ordinary page views and
static asset requests do not produce a request audit record.

## Lifecycle events

- `account.login`: success, incorrect password, account not found, inactive,
  unverified, or invalid email. Existing accounts include their identifier;
  validated email addresses contribute only their domain.
- `account.logout`, `session.created`, `session.expired`, `session.revoked`:
  session creation, logout, expiry during access or periodic cleanup, account
  access revocation, password reset and accepted invitations. Session credentials
  are never recorded.
- `account.register`, `account.email_confirmed`, `account.confirmation_resent`,
  `account.invitation_accepted`: registration and confirmation lifecycle.
- `account.password_reset_requested`, `account.password_reset`, `account.token`:
  recovery mail queued, successful reset, or invalid/expired link.
- `account.mail`: SMTP acceptance, failed SMTP attempt, or queued mail discarded
  after its token expires or is revoked. Mail events carry the account and mail
  identifiers and kind. Technical delivery errors remain in maintenance logs.
- `account.tokens_expired`: periodic token cleanup count.
- `access.denied`, `request.csrf`: rejected authentication, authorization, or
  form-CSRF checks.
- `submission.draft_saved`, `submission.submitted`, `submission.edited`,
  `submission.reviewed`: committed submission changes, actor, submission
  identifier, and moderation status where applicable.
- `submission.mail_preference`, `submission.mail`: reviewer notification settings
  and SMTP outcomes.
- `exam_document.access`: authorized exam document requests, including HEAD/range
  requests. This records authorization, not completion of a download.

## Investigating delivery issues

The persistent history can be queried by account, event, or request ID:

```sql
SELECT occurred_at, event, outcome, actor_user_id, user_id, details
FROM audit_events
WHERE user_id = ? OR actor_user_id = ?
ORDER BY id DESC;
```

For a reset, follow `account.password_reset_requested` to `account.mail`, then
`account.password_reset`. `smtp_accepted` means the relay accepted the message;
recipient delivery, quarantine and bounces must be checked with the mail provider
or recipient. A new reset request replaces the previous link, and reset links
expire after one hour. Public responses continue to conceal whether an account
exists.

Audit output starts with deployment of this code; it cannot reconstruct earlier
login failures.
