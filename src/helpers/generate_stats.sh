#!/usr/bin/env bash

usage() {
    cat <<EOF
Usage:
    $0 <sender.log> <receiver.log>

It reads log files from the metron sender and receiver in order to generates CSV
files with the stats and merge both output in one CSV.
EOF
}


log_csv() {
    local input=${1}
    local output=${2}

    echo -n "Processing stats on ${input} ... "
    awk 'BEGIN { counter=0 } /^--STATS/{ printf("%s,", counter); i(for=2; i<=NF; i++) printf("%s,", $i); print $NF; counter++; } END{ print "STATS processed: " counter >"/dev/stderr" }' ${input} > ${output}
}


generate_csv() {
    local sender=${1}
    local receiver=${2}
    local output=${3}

    # First colum acts like id (its the line number)
    echo "Mergin CSV in ${output} ..."
    join -t, -2 2 -2 3 ${sender} ${receiver} > ${output}
}


# Program
if [ "$0" == "${BASH_SOURCE[0]}" ]
then
    if [ $# -ne 3 ]
    then
        usage
        exit 1
    fi
    SENDER=$1
    RECEIVER=$2
    OUTPUT=$3
    SENDER_CSV="$(basename $SENDER_CSV).csv"
    RECEIVER_CSV="$(basename $RECEIVER_CSV).csv"

    log_csv $SENDER $SENDER_CSV
    log_csv $RECEIVER $RECEIVER_CSV
    generate_csv $SENDER_CSV $RECEIVER_CSV $OUTPUT
    exit 0
fi
