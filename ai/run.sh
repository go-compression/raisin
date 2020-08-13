exec env LD_LIBRARY_PATH=$PWD/engine /usr/bin/python3 -x test.py "$@"
# This ^ sets the LD_LIBRARY_PATH to import the correct shared objects in the engine directory