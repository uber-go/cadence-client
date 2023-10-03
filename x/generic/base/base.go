/*
Package base defines common shared types between many x/generics APIs.

BY USING ANY PACKAGE IN X YOU ACCEPT THAT EVERYTHING MAY CHANGE, AND YOU ARE
QUITE LIKELY TO NOT GET ANY SUPPORT AT ALL.  Use with caution.

This package MUST be a dead-end for x/generic purposes, and should generally
avoid importing other packages where possible, as it will influence build time
for nearly all other packages.
*/
package base

/*
SerializationFriendly defines a... placeholder of sorts, to mark that this
type is intended for serialization.

Though this cannot be enforced at compile time, you are STRONGLY recommended
to use ONLY struct-backed types or something similar, as nearly all
serialization protocols allow adding fields to them.

It primarily exists because Go does not have any way to refer to a type that
is backed by a ~struct{...} with arbitrary contents.  Only other primitives
and interfaces, neither of which are acceptable.

If Go gains the ability to define "this must be a struct" or "this must not
be a primitive or interface type", this type constraint MAY be modified to
use it.
*/
type SerializationFriendly interface{}
