#!/bin/bash

# Define the file path and the URL for downloading
FILE_PATH="$HOME/.nix-profile/share/bash-completion/completions/git-prompt.sh"
DOWNLOAD_URL="https://raw.githubusercontent.com/git/git/master/contrib/completion/git-prompt.sh"

# Check if the file exists
if [ -f "$FILE_PATH" ]; then
    echo "git-prompt.sh already exists at $FILE_PATH"
else
    echo "git-prompt.sh not found. Downloading..."
    # Create the directory if it doesn't exist
    mkdir -p "$(dirname "$FILE_PATH")"
    # Download the file
    curl -o "$FILE_PATH" "$DOWNLOAD_URL"
    echo "git-prompt.sh has been downloaded to $FILE_PATH"
fi
