#!/bin/zsh
set -u

# Run from the folder this file is in
cd "${0:A:h}"

echo "Pals GemFlows"
echo ""
echo "Starting..."
echo ""

./pals-gemflows run

status=$?

if [[ $status -ne 0 ]]; then
	echo ""
	echo "Pals GemFlows exited with an error (code $status)."
	echo "If this is your first run, make sure you created a .env file with GEMINI_API_KEY."
fi

echo ""
echo "Done. Press Enter to close."
read -r
