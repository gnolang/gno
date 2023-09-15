# VRF

For VRF 0.1 for Gnoland, we will use following mechanism.

- VRF data feeders are available and only those can feed data
- Multiple data will be feeded into the realm by multiple feeders
- All of the values provided from those feeders will be combined to generate random data (This will make the random secure enough e.g. when 1 feeder is attacked for server attack - just need at least one trustworthy data feeder)

## Timing

Random data will only be fulfilled up-on the request.
That way, noone knows what will be written at the time of requesting randomness.
