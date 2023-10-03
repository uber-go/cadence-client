/*
Package internal (and subdirs) contains internal-only implementation details for
the x/generic APIs.

Guard the exposed APIs *religiously*.  Do not expose anything that does not
STRICTLY need to be exposed.

This includes type aliases.  Do not ever user them to expose internals.  Wrap
and re-define them in the public API instead.

Internally, do whatever is readable and convenient, it's fine.
*/
package internal
