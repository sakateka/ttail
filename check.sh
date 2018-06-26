vgo build -v -o ttail cmd/ttail.go || exit 1

FIRST=true
TIME=${1:-10s}
CONTEXT=${2:-2}
LOG=${3:-test.log}

SED_RE='s/(^[^\t]+)\t.*\t(timestamp=[^\t]+).*\t(request_id=[0-9a-f.]+)\t.*/    \1 \2 \3/'

while read; do
    if $FIRST; then
        if echo "${REPLY}"|grep -q -e '^>>>' -e '{'; then
            echo "${REPLY}"
            continue
        fi
        echo "First line"
        echo "${REPLY}"
        echo
        echo

        id="$(echo "$REPLY"|grep -Po '\trequest_id=\K[0-9a-f.]+\t')"
        echo "Lines at begin edge (cuted and grep -C$CONTEXT)"
        fgrep -B$CONTEXT $id $LOG|\
            sed -r "$SED_RE;\$s/^   />>>/"
        fgrep -A$CONTEXT $id $LOG|sed -r "1d;$SED_RE"
        FIRST=false
    else
        break
    fi
done < <(./ttail -d -l -n $TIME $LOG 2>&1)|tr '\t' ' '|cut -c-$(tput cols)

rm -v ttail
