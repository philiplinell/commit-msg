#!/bin/bash

NEW_REPO_NAME=$1

# Check if a new repository name was provided
if [ -z "$NEW_REPO_NAME" ]
then
    echo "Error: No repository name provided."
    echo "Usage: ./init.sh <new-repo-name>"
    exit 1
fi

# Replace module path in go.mod
sed -i '' "s|/go-template|/${NEW_REPO_NAME}|g" go.mod

# Replace go-template in makefile
sed -i '' "s|/go-template|/${NEW_REPO_NAME}|g" Makefile

# Install hooks
make install-hooks

rm init.sh

echo "Initialization completed for ${NEW_REPO_NAME}"

