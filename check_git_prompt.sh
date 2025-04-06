#!/bin/bash

# Define the file path and the URL for downloading
FILE_PATH="$HOME/.nix-profile/share/bash-completion/completions/git-prompt.sh"
DIRECTORY_PATH="$(dirname "$FILE_PATH")"
DOWNLOAD_URL="https://raw.githubusercontent.com/git/git/master/contrib/completion/git-prompt.sh"

ls -l /home/runner/.local/state/nix/profiles/profile/
# Check if the directory exists
if [ -d "$DIRECTORY_PATH" ]; then
    echo "Directory $DIRECTORY_PATH already exists."
else
    echo "Directory $DIRECTORY_PATH not found. Creating it..."
    mkdir -p "$DIRECTORY_PATH"
fi

# Check if the file exists
if [ -f "$FILE_PATH" ]; then
    echo "git-prompt.sh already exists at $FILE_PATH"
else
    echo "git-prompt.sh not found. Downloading..."
    # Download the file
    curl -o "$FILE_PATH" "$DOWNLOAD_URL"
    echo "git-prompt.sh has been downloaded to $FILE_PATH"
fi
