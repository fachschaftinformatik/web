# Event image presentation

Reviewed 18 October 2026 against the GOV.UK Design System, W3C WAI and USWDS.

## Decision

An event image supports identifying the event and inspecting its original
artwork. The event title, dates, venue and description remain ordinary HTML.
Do not require readers to decode a poster to understand the event.

On detail pages use one reading column, in this order:

1. Title and publication date.
2. Event time and venue.
3. Optional image with a full-size link.
4. Description, timetable and related links.

This replaces a fixed 240px image sidebar which narrowed the description and
made portrait posters difficult to inspect. The image stays in the same DOM
position on every screen size. Preserve its aspect ratio and show the entire
image, without a crop, overlay, decorative frame or lightbox dependency.
Its current bounds (30rem wide, at most 36rem or 70vh high) are project layout
choices, not dimensions prescribed by GOV.UK. The detail text has a bounded
reading measure rather than sharing a narrow column with the poster.

List cards show the whole event image within their consistent thumbnail area.
The title provides the link and accessible name; redundant card images have
empty alt attributes. Images on detail pages retain their editorial alt text.
Imported automatic descriptions are source material, not a substitute for a
concise editorial description or the readable event information.

## Sources and rationale

- [GOV.UK images](https://design-system.service.gov.uk/styles/images/): use
  images for a real user need, keep essential information in written content,
  and choose alt text according to purpose. No specific event-poster placement
  is prescribed.
- [GOV.UK layout](https://design-system.service.gov.uk/styles/layout/): start
  with a single column and constrain text line lengths. Nested columns are not
  a requirement simply because an image exists.
- [W3C WAI images](https://www.w3.org/WAI/tutorials/images/): distinguish
  informative, decorative and functional images; text-heavy images need an
  appropriate text equivalent.
- [USWDS cards](https://designsystem.digital.gov/components/card/): cards are
  entry points to related content; images must work at different sizes and
  should help distinguish the items. Their layout is not the article layout.

The GOV.UK publishing platform's separate image-upload guidance has platform
specific rules. Use the Design System and WAI purpose-based alt-text guidance
for this custom application, rather than copying publisher-field rules.

## Publication date

Keep `Veröffentlicht am …` directly below the page title as small, secondary
text with a semantic `time` element. The event's actual date and time remain
in the separate `Termin` block. Publication metadata is not a poster caption
or an event-list date. Show it once, only when the original publication date
is known, without inventing a date for records that lack one.

[GOV.UK metadata](https://components.publishing.service.gov.uk/component-guide/metadata)
supports explicitly labelled publication dates. The
[detailed-guide template](https://design-guide.publishing.service.gov.uk/frontend-templates/detailed-guide/)
uses both introductory metadata and a later published-dates component. These
demonstrate available patterns, not a mandated event-page location. Under-title
placement is our choice to keep announcement context near its title without
mixing it into the event schedule.
