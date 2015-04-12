/*
raa is a file container, similar to tar or zip, focused on allowing
constant-time random file access with linear memory consumption increase.

The library implements a very similar API to the go os package, allowing full
control over,and low level acces to the contained files. raa is based on boltdb,
a low-level key/value database for Go.
*/
package raa
