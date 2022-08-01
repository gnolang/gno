All the files in the ./files directory are meant to be those derived/borrowed from Yaegi, Apache2.0.
The tests files in ./files2 are derived from ./files. The difference is that:

 * ./files/* imports use the NativeType and NativeValue native compat system. This tests the native system.
 * ./files2/* don't use Native*, but rely on the ported stdlibs/**/* standard libraries written in Gno.

Generally, when adding new tests to one folder, it should be added to the other.
