#!/bin/sh
# Vendor pinned GOV.UK, HTMX and TinyMCE assets with curl, tar and Dart Sass.
set -eu

version=6.6.0
root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
assets="$root/view/assets"
temp=$(mktemp -d)
trap 'rm -rf "$temp"' EXIT HUP INT TERM

curl --fail --location --silent --show-error \
    "https://registry.npmjs.org/govuk-frontend/-/govuk-frontend-$version.tgz" \
    --output "$temp/govuk.tgz"
tar -xzf "$temp/govuk.tgz" -C "$temp"
sass --no-source-map --style=compressed --quiet-deps \
    --load-path="$temp/package/dist" \
    "$root/view/styles/govuk.scss" "$assets/govuk.css"
cp "$temp/package/dist/govuk/govuk-frontend.min.js" "$assets/govuk.js"
mkdir -p "$assets/fonts"
install -m 0644 "$temp/package/dist/govuk/assets/fonts/"*.woff2 "$assets/fonts/"
curl --fail --location --silent --show-error \
    "https://raw.githubusercontent.com/alphagov/govuk-frontend/v$version/LICENSE.txt" \
    --output "$assets/GOVUK-LICENSE.txt"

curl --fail --location --silent --show-error \
    https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js \
    --output "$assets/htmx.js"
curl --fail --location --silent --show-error \
    https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/LICENSE \
    --output "$assets/HTMX-LICENSE.txt"

version=8.10.0
language_version=26.9.28
gds_revision=c053cf1c198580723c8768f938cb6146dcd122e8

download() {
    curl --fail --location --silent --show-error "$1" --output "$2"
}

mkdir -p "$temp/core" "$temp/language"
download "https://registry.npmjs.org/tinymce/-/tinymce-$version.tgz" "$temp/core.tgz"
download "https://registry.npmjs.org/tinymce-i18n/-/tinymce-i18n-$language_version.tgz" "$temp/language.tgz"
tar -xzf "$temp/core.tgz" -C "$temp/core"
tar -xzf "$temp/language.tgz" -C "$temp/language"

for path in tinymce.min.js license.md themes/silver/theme.min.js \
    icons/default/icons.min.js models/dom/model.min.js \
    skins/content/default/content.min.css plugins/lists/plugin.min.js; do
    mkdir -p "$assets/tinymce/$(dirname "$path")"
    install -m 0644 "$temp/core/package/$path" "$assets/tinymce/$path"
done
mkdir -p "$assets/tinymce/skins/ui/oxide" "$assets/tinymce/langs8"
for file in "$temp/core/package/skins/ui/oxide/"*.min.css; do
    install -m 0644 "$file" "$assets/tinymce/skins/ui/oxide/"
done
install -m 0644 "$temp/language/package/langs8/de.js" "$assets/tinymce/langs8/de.js"
install -m 0644 "$temp/language/package/LICENSE.md" "$assets/tinymce/LICENSE.md"

mkdir -p "$assets/icons"
for name in bullet-list ordered-list link undo redo; do
    download "https://raw.githubusercontent.com/alphagov/govspeak-visual-editor/$gds_revision/lib/icons/$name.svg" "$assets/icons/$name.svg"
done
download "https://raw.githubusercontent.com/alphagov/govspeak-visual-editor/$gds_revision/LICENCE" "$assets/icons/GDS-LICENCE.txt"

design_revision=a1bfd60d1e0fcc2de52a1d357b934c3098ec62a6
design_url="https://raw.githubusercontent.com/alphagov/govuk-design-system/$design_revision"
mkdir -p "$assets/community"
for name in discuss-and-give-feedback.svg share-your-examples-and-research.svg \
    propose-a-change-to-pages.svg research-interview.png catchup-groupshot.png workshop.png; do
    download "$design_url/src/community/images/$name" "$assets/community/$name"
done
download "$design_url/LICENSE" "$assets/community/GDS-LICENCE.txt"
