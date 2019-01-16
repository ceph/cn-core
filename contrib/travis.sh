#!/bin/bash
set -ex


#############
# VARIABLES #
#############
NB_TAGS=$(git rev-list --tags|wc -l)
# do not request the penultimate tag if they are not tags or just one
if [[ "$NB_TAGS" -gt "1" ]]; then
    PENULTIMATE_TAG=$(git describe --abbrev=0 --tags "$(git rev-list --tags --skip=1 --max-count=1)") # this is n-1 tag
fi

#############
# FUNCTIONS #
#############

function setup_git {
    git config --global user.email "buils@travis-ci.com"
    git config --global user.name "Travis CI"
}

function commit_and_push {
    git commit -s -m "$@"
    git pull origin master --rebase
    git push https://"$GITHUB_TOKEN"@github.com/ceph/cn-core master
}

function commit_spec_file {
    pushd contrib
        ./tune-spec.sh "$PENULTIMATE_TAG" "$TRAVIS_TAG"
        git add cn-core.spec
    popd 2>/dev/null
    commit_and_push "Packaging: Update specfile version to $TRAVIS_TAG"
}

function commit_bash_completion {
    COMPLETION_FILE="contrib/cn_core_completion.sh"

    # Generating the new completion file
    ./cn-core completion > $COMPLETION_FILE

    # Does the completion file being modified
    if git ls-files -m | grep -q "$COMPLETION_FILE"; then
        git add $COMPLETION_FILE
        commit_and_push "contrib: Updating bash completion script"
    fi
}

function commit_changed_readme {
    # we replace the n-1 tag with the last one
    sed -i "s/$PENULTIMATE_TAG/$TRAVIS_TAG/g" README.md
    git add README.md
    commit_and_push "Readme: Bump the new release tag: $TRAVIS_TAG"
}

function compile_cn_core {
    make prepare
    make
}

function docker_build {
    sudo docker build -t cn-core .
}

function run_cn {
    time ./cn cluster start cn-core -i cn-core
}


function get_cn {
    curl -L https://github.com/ceph/cn/releases/download/v2.1.1/cn-v2.1.1-linux-amd64 -o cn
    chmod +x cn
}


########
# MAIN #
########
if [[ "$1" == "compile-run-cn-core" ]]; then
    compile_cn_core
    get_cn
    docker_build
    run_cn
fi

if [[ "$1" == "tag-release" ]]; then
    if [ -n "$TRAVIS_TAG" ]; then
        echo "I'm running on tag $TRAVIS_TAG, let's build a new release!"
        ./contrib/release.sh -g "$GITHUB_TOKEN" -t "$TRAVIS_TAG" -p "$PENULTIMATE_TAG" -b "master"
        git checkout master
        setup_git
        commit_spec_file
        # commit_changed_readme
        # commit_bash_completion # not implemented yet
    else
        echo "Not running on a tag, nothing to do!"
    fi
fi