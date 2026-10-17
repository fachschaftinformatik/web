#!/usr/bin/env python3
"""Explicitly import existing site content into a migrated DB and local blob store."""
import argparse
import hashlib
from pathlib import Path
import sqlite3

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("database", type=Path)
parser.add_argument("blobs", type=Path)
args = parser.parse_args()
root = Path(__file__).resolve().parent
if not args.database.is_file():
    parser.error("database must already exist and have schema migration 13 applied")
with sqlite3.connect(args.database) as db:
    db.execute("SELECT file FROM historical_meeting_documents LIMIT 0")
    # Verify and copy bytes before publishing their metadata. Existing objects
    # must match; interrupted imports can be retried without replacing content.
    for file in sorted((root / "documents").glob("*.pdf")):
        data = file.read_bytes()
        digest = hashlib.sha256(data).hexdigest()
        target = args.blobs / "meetings" / "historical" / digest
        target.parent.mkdir(parents=True, exist_ok=True)
        if target.exists():
            if target.read_bytes() != data:
                raise RuntimeError(f"existing object differs: {target}")
        else:
            with target.open("xb") as out:
                out.write(data)
            target.chmod(0o600)
    # Leave the transaction open until the published metadata has been verified.
    script = (root / "content.sql").read_text()
    db.executescript(script.removesuffix("COMMIT;\n"))
    for file, digest, size in db.execute("SELECT file,sha256,bytes FROM historical_meeting_documents"):
        data = (args.blobs / "meetings" / "historical" / digest).read_bytes()
        if len(data) != size or hashlib.sha256(data).hexdigest() != digest:
            raise RuntimeError(f"document integrity mismatch: {file}")
    db.commit()
print("Imported organization settings and historical documents; existing records preserved.")
