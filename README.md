# Bloom

This is a modified bloom filter implementation focused on supporting
[delta-compression](http://www.eecs.harvard.edu/~michaelm/NEWWORK/postscripts/cbf2.pdf).

Standard bloom filters focus on optimizing the false positive ratio in relationship
to the size (in memory) of the filter. Delta compression extends this tradeoff to
consider the size of transferring the filter on the wire. Fewer hash functions and a
larger memory size can be coupled with standard transport compression algorithms for a
lower overall transport size.

This implementation works with a single hash function, allowing optimization of
networks size based on acceptable false positive rate and memory constraints.

## Representative characteristics:

| Memory size | Items | Compressed size (bytes) | false positive rate |
| --- | --- | --- | --- |
| 0.13MB | 1000 | 2894 | 0.0009 |
| 0.13MB | 5000 | 9551 | 0.0046 |
| 0.13MB | 10000 | 15572 | 0.0093 |
| 0.13MB | 20000 | 25028 | 0.0192 |
| 0.26MB | 1000 | 3242 | 0.0004 |
| 0.26MB | 5000 | 11426 | 0.0024 |
| 0.26MB | 10000 | 18880 |0.0047 |
| 0.26MB | 20000 | 30970 | 0.0098 |
| 0.52MB | 1000 | 3572 | 0.0002 |
| 0.52MB | 5000 | 13400 | 0.0011 |
| 0.52MB | 10000 | 22783 | 0.0028 |
| 0.52MB | 20000 | 37764 | 0.0045 |
| 1.05MB | 1000 | 4132 | 0.0001 |
| 1.05MB | 5000 | 15428 | 0.0005 |
| 1.05MB | 10000 | 26757 | 0.0012 |
| 1.05MB | 20000 | 45602 | 0.0024 |
| 2.10MB | 1000 | 5338 | 0.0001 |
| 2.10MB | 5000 | 17060 | 0.0002 |
| 2.10MB | 10000 | 30733 | 0.0007 |
| 2.10MB | 20000 | 53593 | 0.0012 |
| 4.19MB | 1000 | 7588 | 0.0000 |
| 4.19MB | 5000 | 19361 | 0.0001 |
| 4.19MB | 10000 | 34074 | 0.0002 |
| 4.19MB | 20000 | 61404 | 0.0006 |
