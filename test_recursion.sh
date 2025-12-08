#!/bin/bash

echo "=== Testing Recursion Depth Detection ==="

# Test depth 0 (normal execution)
echo -e "\n1. Testing depth 0 (normal):"
SLEEPSHIP_DEPTH=0 bash -c 'echo "Depth: $SLEEPSHIP_DEPTH"'

# Test depth 1
echo -e "\n2. Testing depth 1:"
SLEEPSHIP_DEPTH=1 bash -c 'echo "Depth: $SLEEPSHIP_DEPTH"'

# Test depth 2
echo -e "\n3. Testing depth 2:"
SLEEPSHIP_DEPTH=2 bash -c 'echo "Depth: $SLEEPSHIP_DEPTH"'

# Test depth 3 (should warn)
echo -e "\n4. Testing depth 3 (at max):"
SLEEPSHIP_DEPTH=3 bash -c 'echo "Depth: $SLEEPSHIP_DEPTH"'

# Test depth 4 (should be blocked)
echo -e "\n5. Testing depth 4 (exceeds max):"
SLEEPSHIP_DEPTH=4 bash -c 'echo "Depth: $SLEEPSHIP_DEPTH"'

echo -e "\n=== Test Complete ==="
