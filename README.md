# record

[`github.com/dal-go/record`](https://pkg.go.dev/github.com/dal-go/record)
defines the small, persistence-neutral vocabulary for identifying records,
carrying record data, and describing changes to records.

It exists so API, rendering, command, and facade layers can depend on a stable
record model without importing database sessions, queries, transactions, or an
adapter. Database access belongs to
[DALgo](https://github.com/dal-go/dalgo): DALgo accepts these values at its
boundary and executes a [`record.Changes`](https://pkg.go.dev/github.com/dal-go/record#Changes)
envelope through [`dal.ApplyChanges`](https://pkg.go.dev/github.com/dal-go/dalgo/dal#ApplyChanges).

The module contains two packages:

- [`record`](https://pkg.go.dev/github.com/dal-go/record) for identity, data,
  envelopes, and change batches.
- [`record/update`](https://pkg.go.dev/github.com/dal-go/record/update) for
  declarative field-update commands.

## API

Every exported type and function is listed below. Links lead to its API
documentation.

### `record`

#### Errors

- Variables: [`ErrNoError`](https://pkg.go.dev/github.com/dal-go/record#ErrNoError),
  [`ErrRecordNotFound`](https://pkg.go.dev/github.com/dal-go/record#ErrRecordNotFound),
  [`ErrRecordExists`](https://pkg.go.dev/github.com/dal-go/record#ErrRecordExists).
- Functions: [`IsNotFound`](https://pkg.go.dev/github.com/dal-go/record#IsNotFound),
  [`IsAlreadyExists`](https://pkg.go.dev/github.com/dal-go/record#IsAlreadyExists).

#### Keys and composite IDs

- Types: [`Key`](https://pkg.go.dev/github.com/dal-go/record#Key),
  [`KeyOption`](https://pkg.go.dev/github.com/dal-go/record#KeyOption), and
  [`FieldVal`](https://pkg.go.dev/github.com/dal-go/record#FieldVal).
- Key constructors: [`NewKeyWithID`](https://pkg.go.dev/github.com/dal-go/record#NewKeyWithID),
  [`NewKeyWithParentAndID`](https://pkg.go.dev/github.com/dal-go/record#NewKeyWithParentAndID),
  [`NewIncompleteKey`](https://pkg.go.dev/github.com/dal-go/record#NewIncompleteKey),
  [`NewKeyWithFields`](https://pkg.go.dev/github.com/dal-go/record#NewKeyWithFields), and
  [`NewKeyWithOptions`](https://pkg.go.dev/github.com/dal-go/record#NewKeyWithOptions).
  A non-empty string id containing the reserved `%` panics (the two direct
  constructors) or is returned as an error satisfying
  `errors.Is(err, ErrInvalidStringID)` (`WithKeyID`, via `NewKeyWithOptions`);
  an empty string id stays legal and denotes an incomplete key.
- Key options: [`WithKeyID`](https://pkg.go.dev/github.com/dal-go/record#WithKeyID),
  [`WithStringID`](https://pkg.go.dev/github.com/dal-go/record#WithStringID),
  [`WithIntID`](https://pkg.go.dev/github.com/dal-go/record#WithIntID),
  [`WithFields`](https://pkg.go.dev/github.com/dal-go/record#WithFields), and
  [`WithParentKey`](https://pkg.go.dev/github.com/dal-go/record#WithParentKey).
- Key helpers: [`EscapeID`](https://pkg.go.dev/github.com/dal-go/record#EscapeID),
  [`UnescapeID`](https://pkg.go.dev/github.com/dal-go/record#UnescapeID) (its
  exact inverse), and [`EqualKeys`](https://pkg.go.dev/github.com/dal-go/record#EqualKeys).
- Methods: [`Key.String`](https://pkg.go.dev/github.com/dal-go/record#Key.String)
  (never panics; an incomplete id prints as an empty segment, e.g. `users/`),
  [`Key.CollectionPath`](https://pkg.go.dev/github.com/dal-go/record#Key.CollectionPath),
  [`Key.Level`](https://pkg.go.dev/github.com/dal-go/record#Key.Level),
  [`Key.Parent`](https://pkg.go.dev/github.com/dal-go/record#Key.Parent),
  [`Key.Collection`](https://pkg.go.dev/github.com/dal-go/record#Key.Collection),
  [`Key.Validate`](https://pkg.go.dev/github.com/dal-go/record#Key.Validate)
  (structure and characters, not completeness; `nil` for an otherwise-valid
  incomplete key),
  [`Key.Equal`](https://pkg.go.dev/github.com/dal-go/record#Key.Equal), and
  [`FieldVal.Validate`](https://pkg.go.dev/github.com/dal-go/record#FieldVal.Validate).

#### Path grammar

The DALgo path grammar (`collection/id/collection/id/…`, shared by
`Key.String()`, `dal.CollectionRef.Path()` and `dbschema.SchemaPath` in
`github.com/dal-go/dalgo`) is defined once here, next to `EscapeID`.

- Errors: [`ErrInvalidPath`](https://pkg.go.dev/github.com/dal-go/record#ErrInvalidPath),
  [`ErrInvalidCollectionName`](https://pkg.go.dev/github.com/dal-go/record#ErrInvalidCollectionName),
  [`ErrMalformedIDSegment`](https://pkg.go.dev/github.com/dal-go/record#ErrMalformedIDSegment), and
  [`ErrReservedKeySegment`](https://pkg.go.dev/github.com/dal-go/record#ErrReservedKeySegment)
  (the reserved `{field=value,...}` composite-key grammar; see its doc comment
  for how a caller that cannot import `dalgo`'s `dal.ErrNotSupported`, such as
  this package, still lets `dbschema` satisfy both sentinels).
- Segment splitting: [`SplitPath`](https://pkg.go.dev/github.com/dal-go/record#SplitPath)
  (a leading `/` is optional and never emitted) and
  [`PathAddressesCollection`](https://pkg.go.dev/github.com/dal-go/record#PathAddressesCollection)
  (odd segment count = collection, even = record).
- Names: [`ValidateCollectionName`](https://pkg.go.dev/github.com/dal-go/record#ValidateCollectionName)
  and [`ValidPlaceholderName`](https://pkg.go.dev/github.com/dal-go/record#ValidPlaceholderName)
  (the identifier rule shared with `dalgo/access`'s `Capture` name).
- Id classification: [`IDSegmentKind`](https://pkg.go.dev/github.com/dal-go/record#IDSegmentKind)
  (`ConcreteIDSegment`, `PlaceholderIDSegment`) and
  [`ClassifyIDSegment`](https://pkg.go.dev/github.com/dal-go/record#ClassifyIDSegment).

#### Record envelopes

- Types: [`Record`](https://pkg.go.dev/github.com/dal-go/record#Record),
  [`WithID`](https://pkg.go.dev/github.com/dal-go/record#WithID), and
  [`DataWithID`](https://pkg.go.dev/github.com/dal-go/record#DataWithID).
- Constructors: [`NewRecord`](https://pkg.go.dev/github.com/dal-go/record#NewRecord),
  [`NewRecordWithData`](https://pkg.go.dev/github.com/dal-go/record#NewRecordWithData),
  [`NewRecordWithIncompleteKey`](https://pkg.go.dev/github.com/dal-go/record#NewRecordWithIncompleteKey),
  [`NewRecordWithoutKey`](https://pkg.go.dev/github.com/dal-go/record#NewRecordWithoutKey),
  [`NewWithID`](https://pkg.go.dev/github.com/dal-go/record#NewWithID), and
  [`NewDataWithID`](https://pkg.go.dev/github.com/dal-go/record#NewDataWithID).
- Helper: [`AnyRecordWithError`](https://pkg.go.dev/github.com/dal-go/record#AnyRecordWithError).
- Method: [`WithID.String`](https://pkg.go.dev/github.com/dal-go/record#WithID.String).

#### Change envelopes and data mapping

- Types: [`Updates`](https://pkg.go.dev/github.com/dal-go/record#Updates) and
  [`Changes`](https://pkg.go.dev/github.com/dal-go/record#Changes).
- `Changes` methods: [`RecordsToInsert`](https://pkg.go.dev/github.com/dal-go/record#Changes.RecordsToInsert),
  [`QueueForInsert`](https://pkg.go.dev/github.com/dal-go/record#Changes.QueueForInsert), and
  [`Reset`](https://pkg.go.dev/github.com/dal-go/record#Changes.Reset).
- Mapping functions: [`DataToMap`](https://pkg.go.dev/github.com/dal-go/record#DataToMap)
  and [`MapToData`](https://pkg.go.dev/github.com/dal-go/record#MapToData).

### `record/update`

- Types: [`FieldPath`](https://pkg.go.dev/github.com/dal-go/record/update#FieldPath)
  and [`Update`](https://pkg.go.dev/github.com/dal-go/record/update#Update).
- Functions: [`ByFieldName`](https://pkg.go.dev/github.com/dal-go/record/update#ByFieldName),
  [`ByFieldPath`](https://pkg.go.dev/github.com/dal-go/record/update#ByFieldPath),
  [`DeleteByFieldName`](https://pkg.go.dev/github.com/dal-go/record/update#DeleteByFieldName), and
  [`DeleteByFieldPath`](https://pkg.go.dev/github.com/dal-go/record/update#DeleteByFieldPath).
- Values: [`DeleteField`](https://pkg.go.dev/github.com/dal-go/record/update#DeleteField)
  and [`ServerTimestamp`](https://pkg.go.dev/github.com/dal-go/record/update#ServerTimestamp).
