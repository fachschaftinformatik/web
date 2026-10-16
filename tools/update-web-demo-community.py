#!/usr/bin/env python3
"""Vendor the GOV.UK Community illustrations used by the local prototype."""
from pathlib import Path
from urllib.request import urlopen

REVISION = "a1bfd60d1e0fcc2de52a1d357b934c3098ec62a6"
BASE = f"https://raw.githubusercontent.com/alphagov/govuk-design-system/{REVISION}"
DEST = Path(__file__).resolve().parents[1] / "internal/webdemo/handler/http/assets/community-images"


def main():
    DEST.mkdir(parents=True, exist_ok=True)
    for name in (
        "discuss-and-give-feedback.svg",
        "share-your-examples-and-research.svg",
        "propose-a-change-to-pages.svg",
        "catchup-groupshot.png",
        "workshop.png",
    ):
        with urlopen(f"{BASE}/src/community/images/{name}", timeout=30) as response:
            content = response.read()
        (DEST / name).write_bytes(content)
    with urlopen(f"{BASE}/LICENSE", timeout=30) as response:
        (DEST / "GDS-LICENCE.txt").write_bytes(response.read())


if __name__ == "__main__":
    main()
