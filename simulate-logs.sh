#!/bin/bash

# Script to simulate live log updates for testing the TUI

echo "Starting log simulation..."

# Function to append log entries
append_tskv_log() {
    local file=$1
    local counter=$2
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S")
    echo -e "\ttimestamp=${timestamp}\tlevel=info\tmsg=Live update ${counter}" >> "$file"
}

append_java_log() {
    local file=$1
    local counter=$2
    local timestamp=$(date -u +"%Y-%m-%d %H:%M:%S")
    echo "${timestamp} [INFO] Live update ${counter} from Java application" >> "$file"
}

# Create test logs directory if it doesn't exist
mkdir -p test-logs

counter=1
while true; do
    # Update app1.log (TSKV format)
    append_tskv_log "test-logs/app1.log" $counter
    
    # Update app2.log (Java format)  
    append_java_log "test-logs/app2.log" $counter
    
    echo "Added log entry #${counter}"
    counter=$((counter + 1))
    
    # Wait 2 seconds between updates
    sleep 0.2
done
