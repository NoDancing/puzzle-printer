#!/usr/bin/env python3
import subprocess
import sys
from datetime import date, timedelta
from pathlib import Path

OUT_DIR = Path.home() / "projects/puzzle-site/pdfs"
OUT_DIR.mkdir(parents=True, exist_ok=True)

PUZZLE_PRINTER = Path.home() / "go/bin/puzzle-printer"

start = date.today() - timedelta(days=120)
end = date.today()

current = start
while current <= end:
    out = OUT_DIR / f"crossword-{current}.pdf"
    if out.exists():
        print(f"Skip {current} (already exists)")
    else:
        result = subprocess.run(
            [str(PUZZLE_PRINTER), "--date", str(current), "--output", str(out)],
            capture_output=True, text=True
        )
        if result.returncode == 0:
            print(f"Saved {current}")
        else:
            msg = (result.stderr or result.stdout).strip()
            print(f"FAIL  {current}: {msg}", file=sys.stderr)
    current += timedelta(days=1)
