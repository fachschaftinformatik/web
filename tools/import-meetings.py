#!/usr/bin/env python3
"""Import the supplied FSV export, preserving source bytes and provenance."""
import hashlib
import json
from pathlib import Path
import sys
from zipfile import ZipFile

root = Path(__file__).resolve().parents[1]
assets = root / "internal/handler/http/assets/meetings"
catalog = root / "internal/meetings/catalog.json"
records = {}


def add(source, date, number, kind, note="", year=None):
    key = date or str(year)
    meeting = records.setdefault(key, {
        "year": year or int(date[:4]), "date": date, "number": number, "documents": []
    })
    meeting["documents"].append({"source": source, "kind": kind, "note": note})


dates = {1: "04-02", 2: "04-16", 3: "05-14", 4: "06-18", 6: "07-29", 7: "08-19", 8: "09-09", 9: "09-24"}
for number, day in dates.items():
    month, d = day.split("-")
    note = ""
    if number in (1, 2):
        note = "Die Einladung nennt 2025; Ausstellungsdatum und zugehöriges Protokoll nennen 2026."
    if number == 2:
        note += " Die Überschrift nennt abweichend die 1. Sitzung."
    add(f"Sitzungseinladungen 2026/Einladung zur {number}. Sitzung am {d}.{month}.pdf", f"2026-{day}", number, "Einladung", note)
    if number != 4:
        add(f"Protokolle 2026/FSV Sitzungsprotokoll {number}.pdf", f"2026-{day}", number, "Protokoll")

for number, day in enumerate(["03-14", "04-11", "05-08", "06-07", "07-26", "10-25", "11-15", "12-10"], 1):
    add(f"Sitzungseinladungen 2024/Einladung {number}.pdf", f"2024-{day}", number, "Einladung")
    if number <= 4:
        note = "Das Protokoll nennt im Begrüßungstext abweichend die 1. Sitzung." if number == 2 else ""
        add(f"Protokolle 2024/FSV Sitzungsprotokoll {number}.pdf", f"2024-{day}", number, "Protokoll", note)

add("Sitzungseinladungen 2025/Einladung 9.pdf", "", 9, "Einladung",
    "Datum uneindeutig: Die Einladung nennt den 23.01.2024, ist aber vom 07.01.2025 und im Ordner 2025 abgelegt.", year=2025)

for number, day in {2: "04-13", 3: "05-18", 4: "07-13", 5: "08-31", 6: "10-26", 7: "11-23", 8: "12-21"}.items():
    add(f"Archiv/Einladung {number}.pdf", f"2022-{day}", number, "Einladung")
for number, day in {1: "03-16", 3: "05-18", 4: "07-13", 5: "08-31", 6: "10-26"}.items():
    note = "Die Dokumentüberschrift nennt 2022, der Begrüßungstext abweichend 2021." if number in (1, 3, 4) else ""
    add(f"Archiv/FSV Sitzungs Protokoll {number}.pdf", f"2022-{day}", number, "Protokoll", note)
add("Archiv/Einladung 9.pdf", "2023-01-18", 9, "Einladung")

for name, date, number in [
    ("1", "2020-03-06", 1), ("2_1", "2020-06-10", 2),
    ("3_1", "2020-11-13", 3), ("4_1", "2020-12-11", 4), ("5_1", "2021-01-22", 5),
]:
    add(f"Archiv/Einladung {name}.pdf", date, number, "Einladung")
for name in ["1_1", "3_1"]:
    add(f"Archiv/FSV Sitzungs Protokoll {name}.pdf", "2020-03-06", 1, "Protokoll",
        "Die erste Überschrift und der Begrüßungstext nennen den 06.03.2020, eine weitere Überschrift abweichend den 22.01.2020.")
add("Archiv/FSV Sitzungs Protokoll 4_1.pdf", "2020-11-13", 3, "Protokoll",
    "Der Quelldateiname nennt 4; im Dokument steht die 3. Sitzung vom 13.11.2020.")

assets.mkdir(parents=True, exist_ok=True)
catalog.parent.mkdir(parents=True, exist_ok=True)
with ZipFile(sys.argv[1]) as archive:
    mapped = {d["source"] for m in records.values() for d in m["documents"]}
    supplied = {n for n in archive.namelist() if n.lower().endswith(".pdf")}
    if mapped != supplied:
        raise ValueError(f"Unmapped: {supplied - mapped}; missing: {mapped - supplied}")
    for meeting in records.values():
        unique = {}
        for doc in meeting["documents"]:
            source = doc.pop("source")
            content = archive.read(source)
            digest = hashlib.sha256(content).hexdigest()
            if digest in unique:
                unique[digest]["sources"].append(source)
                continue
            filename = f"{meeting['date'] or meeting['year']}-{doc['kind'].lower()}.pdf"
            doc.update(file=filename, bytes=len(content), sha256=digest, sources=[source])
            (assets / filename).write_bytes(content)
            unique[digest] = doc
        meeting["documents"] = list(unique.values())
catalog.write_text(json.dumps(sorted(records.values(), key=lambda m: (m["year"], m["date"]), reverse=True), ensure_ascii=False, indent=2) + "\n")
print(f"Imported {len(supplied)} sources as {sum(len(m['documents']) for m in records.values())} PDFs for {len(records)} meetings.")
