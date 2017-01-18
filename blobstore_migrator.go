package goblob

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/workpool"

	"github.com/pivotalservices/goblob/blobstore"
)

var (
	buckets = []string{"cc-buildpacks", "cc-droplets", "cc-packages", "cc-resources"}
)

// BlobstoreMigrator moves blobs from one blobstore to another
type BlobstoreMigrator interface {
	Migrate(dst blobstore.Blobstore, src blobstore.Blobstore) error
}

type blobstoreMigrator struct {
	pool         *workpool.WorkPool
	blobMigrator BlobMigrator
	skip         map[string]struct{}
}

func NewBlobstoreMigrator(
	pool *workpool.WorkPool,
	blobMigrator BlobMigrator,
	exclusions []string,
) BlobstoreMigrator {
	skip := make(map[string]struct{})
	for i := range exclusions {
		skip[exclusions[i]] = struct{}{}
	}

	return &blobstoreMigrator{
		pool:         pool,
		blobMigrator: blobMigrator,
		skip:         skip,
	}
}

func (m *blobstoreMigrator) Migrate(dst blobstore.Blobstore, src blobstore.Blobstore) error {
	if src == nil {
		return errors.New("src is an empty store")
	}

	if dst == nil {
		return errors.New("dst is an empty store")
	}

	fmt.Printf("Migrating from %s to %s\n\n", src.Name(), dst.Name())

	stats := &migrateStats{}
	stats.Start()

	var migrateWG sync.WaitGroup
	for _, bucket := range buckets {
		if _, ok := m.skip[bucket]; ok {
			continue
		}

		iterator, err := src.NewBucketIterator(bucket)
		if err != nil {
			return fmt.Errorf("could not create bucket iterator for bucket %s: %s", bucket, err)
		}

		fmt.Print(bucket)

		var bucketWG sync.WaitGroup
		var migrateErrors []error
		for {
			blob, err := iterator.Next()
			if err == blobstore.ErrIteratorDone {
				break
			}

			if err != nil {
				return err
			}

			checksum, err := src.Checksum(blob)
			if err != nil {
				return fmt.Errorf("could not checksum blob: %s", err)
			}

			blob.Checksum = checksum

			migrateWG.Add(1)
			bucketWG.Add(1)
			m.pool.Submit(func() {
				defer bucketWG.Done()
				defer migrateWG.Done()

				if !dst.Exists(blob) {
					err := m.blobMigrator.Migrate(blob)
					if err != nil {
						fmt.Print("x")
						stats.AddFailed()
						migrateErrors = append(migrateErrors, err)
						return
					}
					stats.AddSuccess()
					fmt.Print(".")
				} else {
					stats.AddSkipped()
					fmt.Print("-")
				}
			})
		}

		bucketWG.Wait()
		fmt.Println(" done.")

		if migrateErrors != nil {
			for i := range migrateErrors {
				fmt.Fprintln(os.Stderr, migrateErrors[i])
			}
		}
	}

	migrateWG.Wait()
	stats.Finish()

	fmt.Println(stats)

	return nil
}

type migrateStats struct {
	startTime  time.Time
	finishTime time.Time
	Duration   time.Duration
	Migrated   int64
	Skipped    int64
	Failed     int64
}

func (m *migrateStats) Start() {
	m.startTime = time.Now()
}

func (m *migrateStats) Finish() {
	m.finishTime = time.Now()
	m.Duration = m.finishTime.Sub(m.startTime)
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

(.) Migrated files:    {{.Migrated}}
(-) Already migrated:  {{.Skipped}}
(x) Failed to migrate: {{.Failed}}
`))

	buf := new(bytes.Buffer)
	err := t.Execute(buf, m)
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}
