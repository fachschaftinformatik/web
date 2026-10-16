#!/usr/bin/env python3
"""Create the clearly labelled sample PDF used by the local archive layout."""

from pathlib import Path


def pdf_text(text):
    return text.replace("\\", "\\\\").replace("(", "\\(").replace(")", "\\)")


lines = [
    "Beispiel-Reko",
    "Diese Datei gehört zum lokalen Layoutentwurf des Altklausurenarchivs.",
    "Sie enthält keine echte Klausurrekonstruktion oder Prüfungsaufgaben.",
]
commands = ["BT", "/F1 18 Tf", "50 780 Td"]
for index, line in enumerate(lines):
    if index:
        commands.extend(["/F1 12 Tf", "0 -35 Td"])
    commands.append(f"({pdf_text(line)}) Tj")
commands.append("ET")
stream = "\n".join(commands).encode("cp1252")
objects = [
    b"<< /Type /Catalog /Pages 2 0 R >>",
    b"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
    b"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] "
    b"/Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
    b"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
    f"<< /Length {len(stream)} >>\nstream\n".encode() + stream + b"\nendstream",
]
data = bytearray(b"%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
offsets = [0]
for index, obj in enumerate(objects, 1):
    offsets.append(len(data))
    data.extend(f"{index} 0 obj\n".encode() + obj + b"\nendobj\n")
xref = len(data)
data.extend(f"xref\n0 {len(offsets)}\n0000000000 65535 f \n".encode())
for offset in offsets[1:]:
    data.extend(f"{offset:010d} 00000 n \n".encode())
data.extend(
    f"trailer\n<< /Size {len(offsets)} /Root 1 0 R >>\nstartxref\n{xref}\n%%EOF\n".encode()
)
target = Path(__file__).resolve().parents[1] / "internal/handler/http/assets/archive-example.pdf"
target.write_bytes(data)
print(f"Wrote {target.name} ({len(data)} bytes).")
