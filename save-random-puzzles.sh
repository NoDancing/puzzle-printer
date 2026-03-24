#!/bin/bash

DIR="${1:-$HOME/projects/puzzle-site/pdfs}"
mkdir -p "$DIR"

# Generate 8 unique random dates within the past year
TODAY=$(date +%s)
ONE_YEAR_AGO=$((TODAY - 365 * 86400))

seen=""
count=0

while [ $count -lt 8 ]; do
  OFFSET=$((RANDOM * RANDOM % (365 * 86400)))
  TS=$((ONE_YEAR_AGO + OFFSET))
  DATE=$(date -r $TS +%F)

  if ! echo "$seen" | grep -q "$DATE"; then
    seen="$seen $DATE"
    echo "[$((count + 1))/8] Fetching $DATE..."
    if ~/go/bin/puzzle-printer --date "$DATE" --output "$DIR/crossword-$DATE.pdf"; then
      count=$((count + 1))
    else
      echo "  Skipping $DATE (fetch failed)"
    fi
  fi
done

echo "Done. PDFs saved to $DIR"
