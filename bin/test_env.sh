#!/usr/bin/env sh

# Usage:
#   ./test_env.sh [--multinode] [--teardown]
#
# If --multinode is set, launches node1 and node2. Otherwise, launches node1 only.
# If --teardown is set, deletes all multipass nodes and exits.

set -e

if ! command -v multipass >/dev/null 2>&1; then
  echo "Error: multipass could not be found. Please install multipass first."
  exit 1
fi

MULTINODE=0
TEARDOWN=0
for arg in "$@"
do
  case "$arg" in
    --multinode) MULTINODE=1 ;;
    --teardown) TEARDOWN=1 ;;
  esac
done

if [ "$TEARDOWN" -eq 1 ]; then
  echo "Deleting all multipass nodes..."
  NODES=$(multipass list --format csv | awk -F, 'NR>1 {print $1}')
  if [ -n "$NODES" ]; then
    multipass delete $NODES --purge
  else
    echo "No multipass nodes to delete."
  fi
  multipass list
  exit 0
fi

if [ "$MULTINODE" -eq 1 ]; then
  NODES="node1 node2"
else
  NODES="node1"
fi


if [ "$MULTINODE" -eq 1 ]; then
  MEMORY=2G
else
  MEMORY=4G
fi

for NODE in $NODES; do
  multipass launch --name "$NODE" --cpus 2 --memory $MEMORY --disk 10G
  PASS=${MPS_PASSWORD:-password123}
  multipass exec "$NODE" -- sudo bash -c "echo ubuntu:${PASS} | sudo chpasswd"
  for key in ~/.ssh/*.pub; do
    multipass exec "$NODE" -- bash -c "echo '$(cat "$key")' >> ~/.ssh/authorized_keys"
  done
  echo "Node $NODE ready."
done

multipass info | grep "IPv4"
