package goblob

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mgutz/ansi"
	"github.com/pivotalservices/goblob/blobstore"
)

var red = ansi.ColorFunc("red+b")
var yellow = ansi.ColorFunc("yellow+b")
var green = ansi.ColorFunc("green+b")

type BlobstoreMigrationWatcher interface {
	MigrationDidStart(blobstore.Blobstore, blobstore.Blobstore)
	MigrationDidFinish()

	MigrateBucketDidStart(string)
	MigrateBucketDidFinish()

	MigrateBlobDidFailWithError(error)
	MigrateBlobDidFinish()
	MigrateBlobAlreadyFinished()
}

//go:generate counterfeiter . BlobstoreMigrationWatcher

func NewBlobstoreMigrationWatcher() BlobstoreMigrationWatcher {
	return &blobstoreMigrationWatcher{
		stats:       &migrateStats{},
		errorsMutex: &sync.Mutex{},
	}
}

type blobstoreMigrationWatcher struct {
	stats       *migrateStats
	errors      []error
	errorsMutex *sync.Mutex
}

func (w *blobstoreMigrationWatcher) MigrationDidStart(dst, src blobstore.Blobstore) {
	fmt.Printf("Migrating from %s to %s\n\n", src.Name(), dst.Name())
	w.stats.Start()
}

func (w *blobstoreMigrationWatcher) MigrationDidFinish() {
	w.stats.Finish()
	fmt.Println(w.stats)
	for i := range w.errors {
		fmt.Fprintln(os.Stderr, w.errors[i])
	}
}

func (w *blobstoreMigrationWatcher) MigrateBucketDidStart(bucket string) {
	fmt.Printf("%s ", bucket)
}

func (w *blobstoreMigrationWatcher) MigrateBucketDidFinish() {
	fmt.Println(" done.")
}

func (w *blobstoreMigrationWatcher) MigrateBlobDidFailWithError(err error) {
	w.errorsMutex.Lock()
	defer w.errorsMutex.Unlock()
	w.stats.AddFailed()
	w.errors = append(w.errors, err)
	fmt.Print(red("."))
}

func (w *blobstoreMigrationWatcher) MigrateBlobDidFinish() {
	w.stats.AddSuccess()
	fmt.Print(green("."))
}

func (w *blobstoreMigrationWatcher) MigrateBlobAlreadyFinished() {
	w.stats.AddSkipped()
	fmt.Print(yellow("."))
}

type migrateStats struct {
	startTime time.Time
	Duration  time.Duration
	Migrated  int64
	Skipped   int64
	Failed    int64
}

func (m *migrateStats) Start() {
	m.startTime = time.Now()
}

func (m *migrateStats) Finish() {
	m.Duration = time.Since(m.startTime)
}

func (m *migrateStats) AddSuccess() {
	atomic.AddInt64(&m.Migrated, 1)
}

func (m *migrateStats) AddSkipped() {
	atomic.AddInt64(&m.Skipped, 1)
}

func (m *migrateStats) AddFailed() {
	atomic.AddInt64(&m.Failed, 1)
}

func (m *migrateStats) String() string {
	t := template.Must(template.New("stats").Parse(`
Took {{.Duration}}

Migrated files:    {{.Migrated}}
Already migrated:  {{.Skipped}}
Failed to migrate: {{.Failed}}
`))

	buf := new(bytes.Buffer)
	err := t.Execute(buf, m)
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}
