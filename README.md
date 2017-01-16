# goblob

`goblob` is a tool for migrating Cloud Foundry blobs from one blobstore to
another. Presently it only supports migrating from an NFS blobstore to an
S3-compatible one.

## Installing

Download the [latest release](https://github.com/pivotalservices/goblob/releases/latest).

### Install from source

Requirements:

* [glide](https://github.com/masterminds/glide)
* [go](https://golang.org)

```
mkdir -p $GOPATH/src/github.com/pivotalservices/goblob
git clone git@github.com:pivotalservices/goblob.git $GOPATH/src/github.com/pivotalservices/goblob
cd $GOPATH/src/github.com/pivotalservices/goblob
glide install
GOARCH=amd64 GOOS=linux go install github.com/pivotalservices/goblob/cmd/goblob
```

## Usage

`goblob migrate [OPTIONS]`

### Options

* `concurrent-uploads`: Number of concurrent uploads (default: 20)
* `exclude`: Directory to exclude (may be given more than once)

#### NFS-specific Options

* `blobstore-path`: The path to the root of the NFS blobstore, e.g. /var/vcap/store/shared

#### S3-specific Options

* `s3-endpoint`: The endpoint of the S3-compatible blobstore
* `s3-accesskey`: The access key to use with the S3-compatible blobstore
* `s3-secretkey`: The secret key to use with the S3-compatible blobstore
* `s3-region`: The region to use with the S3-compatible blobstore
* `cf-identifier`: The suffix appended to the blobstore directories.
* `use-multipart-uploads`: Whether to use multi-part uploads
* `disable-ssl`: Whether to disable SSL when uploading blobs

## Post-migration Tasks

- Update OpsManager File Storage Config to point at S3 blobstore using buckets (cc-buildpacks-<uniqueid>, cc-droplets-<uniqueid>, cc-packages-<uniqueid>, cc-resources-<uniqueid>)
- Click `Apply Changes` in OpsManager
- Once changes are applied, re-run `goblob` to migrate any files which were created after the initial migration
- Validate apps can be restaged and pushed
- Update NFS resource config to 0 in OpsManager to remove NFS server
- Click `Apply Changes` in OpsManager

## Developing

* Install [Docker](https://www.docker.com/products/docker)
* `docker pull minio/minio`

To run all of the tests in a Docker container:

`./testrunner`

To continually run the tests during development:

* `docker run -p 9000:9000 -e "MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE" -e "MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" minio/minio server /tmp`
* (in a separate terminal) `ginkgo watch -r`
