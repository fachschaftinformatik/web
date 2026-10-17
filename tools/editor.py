#!/usr/bin/env python3
"""Vendor the pinned TinyMCE core, selected plugins and German translation."""
import io
import tarfile
import urllib.request
from pathlib import Path

VERSION = "8.10.0"
LANG_VERSION = "26.9.28"
ROOT = Path(__file__).resolve().parents[1] / "internal/web/assets/tinymce"
PLUGINS = {"lists"}

def unpack(url, accept):
    with urllib.request.urlopen(url) as response:
        data = response.read()
    with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as archive:
        for entry in archive.getmembers():
            path = Path(entry.name)
            relative = Path(*path.parts[1:])
            if entry.isfile() and ".." not in relative.parts and accept(str(relative)):
                target = ROOT / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(archive.extractfile(entry).read())

unpack(f"https://registry.npmjs.org/tinymce/-/tinymce-{VERSION}.tgz", lambda p:
       p in {"tinymce.min.js", "license.md"} or
       p in {"themes/silver/theme.min.js", "icons/default/icons.min.js", "models/dom/model.min.js"} or
       p == "skins/content/default/content.min.css" or
       (p.startswith("skins/ui/oxide/") and p.endswith(".min.css")) or
       any(p == f"plugins/{name}/plugin.min.js" for name in PLUGINS))
unpack(f"https://registry.npmjs.org/tinymce-i18n/-/tinymce-i18n-{LANG_VERSION}.tgz",
       lambda p: p == "langs8/de.js" or "license" in p.lower())
print(f"Vendored TinyMCE {VERSION} and translations {LANG_VERSION} in {ROOT}")

# These are the actual GDS editor UI assets, not an editor-engine dependency.
GDS_REV = "c053cf1c198580723c8768f938cb6146dcd122e8"
icons = ROOT.parent / "icons"
icons.mkdir(exist_ok=True)
for name in ("bullet-list", "ordered-list", "link", "undo", "redo"):
    with urllib.request.urlopen(f"https://raw.githubusercontent.com/alphagov/govspeak-visual-editor/{GDS_REV}/lib/icons/{name}.svg") as response:
        (icons / f"{name}.svg").write_bytes(response.read())
with urllib.request.urlopen(f"https://raw.githubusercontent.com/alphagov/govspeak-visual-editor/{GDS_REV}/LICENCE") as response:
    (icons / "GDS-LICENCE.txt").write_bytes(response.read())
