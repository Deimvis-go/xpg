package pgfw

// func

// NewQuery(ctx, queryText, args...)
// .Exec() (cmdTag, err)
// .ReturnRows().Exec() (cmdTag, rows, err)
// -rows interface
// .Pull()
// .Delegate(puller interface {Pull() row})
// TODO: figure out how to merge into one declaration to avoid manual rows.Close()
// mb .ReturnRows().Into(puller ).Exec() // TODO: type param
// mb .PullRows(puller interface[T]{ Pull(row) (bool, err) }).Exec() // TODO: type param
//  var p puller = pgfw.NewStructRowPuller[MyStruct](pgfw.PullOneRow(), pgfw.ValidateOnFinish(), ...)
//  cmdTag, err := pgfw.NewQuery(ctx, query, args...).PullRows(puller).Exec()
//  // unclear whether err represents puller errors
//  // unclear how puller should bufferize into struct? most likely it should accept slice (/array) in constructor
//  // too long declaration
// TODO: somehow deal with that cmd tag is returned only when rows are read and closed

// v2
// NewQuery(ctx, queryText, args...)
// .Exec() (cmdTag, err)
// .ReturnRows().Exec() (cursor, err) // res.CmdTag() (string, ok bool)
// .ReturnRows().Into(puller).Exec() (cmdTag, err)
// TODO: maybe require puller interface to return result slice so we can propagate it to return values? but puller may live not over slice, so no
// .ReturnRows().IntoSlice(slicePuller).Exec() (cmdTag, err)
// e.g. .IntoSlice(pgfw.NewSlicePuller())
// .ReturnRows().MapIntoNewSlice(mapFn).Exec() (cmdTag, err)
