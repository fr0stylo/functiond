#!/bin/bash
# This script cleans up orphaned CNI chains
#
#chains=$(sudo iptables -t nat -L | grep 'CNI-' | awk '{print $1}')
#for chain in $chains; do
#  iptables -t nat -F $chain
#  iptables -t nat -X $chain
#done
#
#echo "CNI iptables chains cleaned up."


ctr -n example tasks kill -s SIGKILL redis-server
ctr -n example container rm node-server-VYbL
ctr -n example snapshots rm node-server-snapshotcommited
