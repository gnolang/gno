#!/bin/sh

main() {
    cd "$(dirname "$0")"
    cd ../..
    fname="$(mktemp --tmpdir gno_file_commits.XXXXXXXXXX.csv)"
    for file in $(list_gno_files); do
        extract_file_metadata $file
    done > "$fname"
    cat "$fname" | sort_by_date | unique_by_author
}

list_gno_files() {
    # list .gno file in examples/, remove tests and unit tests
    find ./examples -name "*.gno" | grep -v _filetest.gno | grep -v _test.gno | grep -v gno.land/r/demo/tests
}

extract_file_metadata() {
    file=$1
    # get the first commit date of the file
    first_commit_date=$(git log --pretty=format:%ct --follow $file | tail -n 1)
    # get the email of the first contributor of the file
    email=$(git log --mailmap --pretty=format:%aE --follow $file | tail -n 1)
    # print the file name, first commit date, and email
    echo "$first_commit_date,$email,$file"
}

sort_by_date() {
    sort -t, -k1
}

unique_by_author() {
    awk -F, '!seek[$2]++'
}

main
