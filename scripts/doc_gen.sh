#!/bin/bash
# Script that checks the code for errors.

set -e
set -x

function generate_godoc {
    oldpwd=$(pwd)
    rm -rf $1
    mkdir $1
    for i in $(find . -iname 'doc.go' -not -path "*vendor/*"); do
        dir=${i%/*}
        mypath="${oldpwd#*/src/}"
        realdir=$mypath${dir:1}
        package=${realdir##${GOPATH}/src/}
        cd ${dir}
	    package_sub=${package/./_}
	    package_sub=${package_sub////_}
	    touch $1$package_sub.html
	    echo "$1$package_sub.html"
	    godoc -ex -html ${package} > $1$package_sub.html
        cd ${oldpwd}
    done;
}


function compare {
    echo "Generating HTML diff"
    oldpwd=$(pwd)
    for i in $(find . -iname 'doc.go' -not -path "*vendor/*"); do
        dir=${i%/*}
        mypath="${oldpwd#*/src/}"
        realdir=$mypath${dir:1}
        package=${realdir##${GOPATH}/src/}
        cd ${dir}
	    package_sub=${package/./_}
	    package_sub=${package_sub////_}
        echo "prev: $1$package_sub.html"
        echo "latest: $2$package_sub.html"
        go run ${TRAVIS_BUILD_DIR}/scripts/diffhtml.go -prev=$1$package_sub.html -latest=$2$package_sub.html
        cd ${oldpwd}
    done;
}

if [ "${TRAVIS_PULL_REQUEST_BRANCH:-$TRAVIS_BRANCH}" != "master" ]; then
    pull=$(git rev-parse HEAD)
    cd ${TRAVIS_BUILD_DIR}/..
    topdir=$(pwd)
    docspr=${topdir}/docs_pr/
    docsmaster=${topdir}/docs_master/
    cd ${TRAVIS_BUILD_DIR}/
    generate_godoc ${docspr}
    git fetch origin master
    git checkout master
    generate_godoc ${docsmaster}
    git checkout ${pull}
    compare ${docsmaster} ${docspr}
fi