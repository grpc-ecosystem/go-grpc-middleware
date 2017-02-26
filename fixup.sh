#!/bin/bash
# Script that checks the code for errors.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"

function print_real_go_files {
    grep --files-without-match 'DO NOT EDIT!' $(find . -iname '*.go')
}

function generate_markdown {
    echo "Generating markdown"
    oldpwd=$(pwd)
    for i in $(find . -iname 'doc.go'); do
        dir=${i%/*}
        echo "$dir"
        cd ${dir}
        ${GOPATH}/bin/godocdown -heading=Title -o DOC.md
        ln -s DOC.md README.md 2> /dev/null # can fail
        cd ${oldpwd}
    done;
}

function goimports_all {
    echo "Running goimports"
    goimports -l -w $(print_real_go_files)
    return $?
}

generate_markdown
goimports_all
echo "returning $?"