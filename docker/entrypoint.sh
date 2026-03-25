#!/bin/sh
echo "Puzzle printer started. Will run daily at 05:00 ${TZ}."

while true; do
    # Seconds elapsed since midnight (force decimal to avoid octal issues with 08/09)
    H=$(date +%H); M=$(date +%M); S=$(date +%S)
    elapsed=$(( 10#$H * 3600 + 10#$M * 60 + 10#$S ))

    # 5:00 AM = 18000 seconds since midnight
    target=18000

    if [ "$elapsed" -le "$target" ]; then
        sleep_secs=$(( target - elapsed ))
    else
        sleep_secs=$(( 86400 - elapsed + target ))
    fi

    echo "Next run in ${sleep_secs}s (at 05:00)"
    sleep "$sleep_secs"

    echo "Running puzzle-printer at $(date)"
    /usr/local/bin/puzzle-printer || true

    # Sleep past the minute boundary so we don't double-fire
    sleep 65
done
