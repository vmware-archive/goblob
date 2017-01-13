package commands

type GoblobCommand struct {
	Version func() `command:"version" description:"Print version information and exit"`

	Migrate MigrateCommand `command:"migrate" description:"Migrate blobs from one blobstore to another"`
}

var Goblob GoblobCommand
