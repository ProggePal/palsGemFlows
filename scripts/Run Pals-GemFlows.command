#!/bin/sh
set -u

# Run from the folder this file is in (portable across shells).
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

printf "Pals GemFlows\n\n"
printf "Starting...\n\n"

./pals-gemflows run
status=$?

if [ "$status" -ne 0 ]; then
  printf "\nPals GemFlows exited with an error (code %s).\n" "$status"
  printf "If this is your first run, the app will ask for your Gemini API key.\n"
fi

printf "\nDone. Press Enter to close.\n"
read _
