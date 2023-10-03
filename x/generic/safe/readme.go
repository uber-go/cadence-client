/*
Package safe provides APIs that are similar to [go.uber.org/cadence/x/generic] in
most ways, but provide MUCH stronger safety guarantees / implications, at the
cost of some assumptions and significant restrictions.

If these are too strict for some legitimate reason, consider using (in
decreasing safety order):

  - the "normal generic" [go.uber.org/cadence/x/generic] APIs which enforce only
    patterns that are true with most plugins / DataConverters / etc.
  - the "primitive generic" [go.uber.org/cadence/x/generic/bare] APIs which
    require as little as possible while still providing generics.
  - or the "type-unsafe" [go.uber.org/cadence] APIs, which allow nearly anything,
    but do not offer generics.
*/
package safe
