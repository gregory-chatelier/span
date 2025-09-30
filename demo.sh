#!/bin/bash

# A simple script to demonstrate the features of the 'span --spark' command.
# Note: This script assumes 'span' is in the current directory and you are in a bash-compatible shell (like Git Bash).

# --- Utility function to print commands before running them ---
run() {
    echo "$ $*"
    eval "$@"
}

clear
echo "================================================="
echo "     'span --spark' Feature Demonstration     "
echo "================================================="

echo
echo "--- DEMO 1: Basic Sparkline from a finite stream ---"
echo "The interval is detected automatically from the input data."
echo
run 'echo "1 5 22 13 5 17 9" | ./span --spark'
echo

read -p "Press Enter to continue..."
clear

echo "--- DEMO 2: Using a fixed interval [0-100] and color ---"
echo "This is useful for data like percentages, where the scale is known."
echo
run 'echo "0 10 20 30 40 50 60 70 80 90 100" | ./span --spark 0 100 --spark-color=blue'
echo

read -p "Press Enter to continue..."
clear

echo "--- DEMO 3: Real-time sliding window animation (runs for 20 seconds) ---"
echo "Simulates a live data feed. The sparkline has a fixed width and updates in place."
echo "The scale is dynamic, based on the min/max of the values currently in the window."
echo

COMMAND='(
  end=$((SECONDS+20))
  while [ $SECONDS -lt $end ]; do
    echo $(($RANDOM % 100))
    sleep 0.05
  done
) | ./span --spark --spark-width 40 --spark-color green'

echo "$ $COMMAND"
(
  end=$((SECONDS+20))
  while [ $SECONDS -lt $end ]; do
    echo $(($RANDOM % 100))
    sleep 0.05
  done
) | ./span --spark --spark-width 40 --spark-color green

echo

echo
echo "--- DEMO COMPLETE ---"
