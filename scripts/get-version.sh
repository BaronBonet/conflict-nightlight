#!/bin/bash

function main() {
     if [ -n "$1" ] && [ "${1}" = "github" ]; then
         echo "VERSION=$(printVersion) >> $GITHUB_OUTPUT"
     else
         git describe --always
     fi
}

main "$@"
