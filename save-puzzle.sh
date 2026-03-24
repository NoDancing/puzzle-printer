#!/bin/bash
DATE="${1:-$(date +%F)}"
DIR="${2:-$HOME/projects/puzzle-site/pdfs}"
mkdir -p "$DIR"
~/go/bin/puzzle-printer --date "$DATE" --output "$DIR/crossword-$DATE.pdf"
