#!/bin/bash

set -e

LOCKFILE="sing-box.lock"

# Exclusive file lock to prevent multiple daemon script processes
exec 9>"$LOCKFILE"
if ! flock -n 9; then
  echo "Error: Another instance of runSingbox.sh is already running with exclusive lock."
  exit 1
fi

TAIL_PID=""

runSingbox(){
  # Clean up any stale sing-box runs to prevent port collision/multi-run
  pkill -f "./sing-box run" || true
  export enable_deprecated_special_outbounds=true
  
  # Setup physical log redirect and output forwarding to stdout
  touch sing-box.log
  pkill -f "tail -f sing-box.log" || true
  tail -f sing-box.log &
  TAIL_PID=$!
  
  ./sing-box run > sing-box.log 2>&1 &
  tokill=$!
  echo "Sing-Box started with PID: $tokill, Log tail PID: $TAIL_PID"
}

terminateSingbox()
{
  if [ ! -z "$TAIL_PID" ] && kill -0 $TAIL_PID > /dev/null 2>&1; then
    echo "Terminating log tail PID=$TAIL_PID"
    kill $TAIL_PID || true
  fi
  if kill -0 $tokill > /dev/null 2>&1; then
    echo "Terminating singbox PID=$tokill"
    kill $tokill
    while kill -0 $tokill > /dev/null 2>&1; do
      sleep 1
    done
  fi
}

reloadSingbox()
{
  if kill -0 $tokill > /dev/null 2>&1; then
    echo "Reloading singbox via SIGHUP PID=$tokill"
    kill -HUP $tokill
  else
    runSingbox
  fi
}

trap terminateSingbox SIGINT SIGTERM SIGKILL
trap reloadSingbox SIGHUP

runSingbox

while true
do
    sleep 2
    if [ -f "signal" ]; then
        (
            flock -x 200
            if [ -f "signal" ]; then
                signal=$(cat signal)
                echo "Signal received: $signal"
                rm -f signal
                case ${signal} in
                    "stop")
                        terminateSingbox
                        ;;
                    "restart")
                        reloadSingbox
                        ;;
                esac
            fi
        ) 200>signal.lock
    fi

    # Check if sing-box crashed
    if ! kill -0 $tokill > /dev/null 2>&1; then
        if [ "$signal" != "stop" ]; then
            echo "Sing-Box with PID $tokill crashed. Breaking the loop..."
            exit 1
        fi
    fi
done
