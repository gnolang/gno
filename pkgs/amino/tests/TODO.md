* Test case where the native time module is not necessary to import, because we
  use timestamppb.Timestamp.AsTime() instead.  This isn't tested for common.go
due to the usage of []time.Time (JAE: I think).

* Test case of string TypeDef (and other types), as they have different code paths.
