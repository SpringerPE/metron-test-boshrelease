#!/usr/bin/env bash

usage() {
    cat <<EOF
Usage:
    $0 <sender.log> <receiver.log>

It reads log files from the metron sender and receiver in order to generates CSV
files with the stats and merge both output in one CSV.
EOF
}


log_receiver_csv() {
    local input=${1}
    local output=${2}

    echo -n "Processing stats on ${input} -> ${output} ... "
    echo "0,Date,Time,NumCPUs,diodes,Workers,LogsReceived,Errors,Duration,Rate(logs/s),logSenderTotalMessagesRead,grpcSendErrors,dropsondeUnmarshallerLogMessages,dopplerIngress" > "${output}"
    awk 'BEGIN { counter=1 } /^--STATS/{ printf("%s,", counter); for (i=2; i<NF; i++) printf("%s,", $i); print $NF; counter++; } END{ print "STATS processed: " counter-1 >"/dev/stderr" }' "${input}" >> "${output}"
}

log_sender_csv() {
    local input=${1}
    local output=${2}

    echo -n "Processing stats on ${input} -> ${output} ... "
    echo "0,NumCPUs,Workers,Interval(us),TheoreticalRate(logs/s),LogsSent,Errors,Duration,Rate(logs/s)" > "${output}"
    awk 'BEGIN { counter=1 } /^--STATS/{ printf("%s,", counter); for (i=2; i<NF; i++) printf("%s,", $i); print $NF; counter++; } END{ print "STATS processed: " counter-1 >"/dev/stderr" }' "${input}" >> "${output}"
}


generate_csv() {
    local sender=${1}
    local receiver=${2}
    local output=${3}

    # First colum acts like id (its the line number)
    echo "Mergin CSV in ${output} ..."
    join -t, ${sender} ${receiver} > ${output}
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
    SENDER_CSV="$(basename ${SENDER}).csv"
    RECEIVER_CSV="$(basename ${RECEIVER}).csv"

    log_sender_csv $SENDER $SENDER_CSV
    log_receiver_csv $RECEIVER $RECEIVER_CSV
    generate_csv $SENDER_CSV $RECEIVER_CSV $OUTPUT
    exit 0
fi

