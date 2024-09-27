

pkill -f 'build/gnoland'
pkill -f 'build/gnoweb'

# Remove the test directory if it exists
rm -rf gno.land/testdir

# Navigate to the gno.land directory
cd gno.land

# Rebuild gnoland and gnoweb to reflect changes
make build

# Start gnoland and gnoweb
./build/gnoland start -lazy &
sleep 5
./build/gnoweb -bind localhost:8888 &
sleep 2