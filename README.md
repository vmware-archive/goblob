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

The tool is a Golang binary, which must be executed on the NFS VM that you intend to migrate. The only command of the tool is this:

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
* `buildpacks-bucket-name`: The bucket containing buildpacks
* `droplets-bucket-name`: The bucket containing droplets
* `packages-bucket-name`: The bucket containing packages
* `resources-bucket-name`: The bucket containing resources
* `use-multipart-uploads`: Whether to use multi-part uploads
* `disable-ssl`: Whether to disable SSL when uploading blobs
* `insecure-skip-verify`: Skip server SSL certificate verification

## Post-migration Tasks

- If your S3 service uses an SSL certificate signed by your own CA: Before applying changes in Ops Manager to switch to S3, make sure the root CA cert that signed the endpoint cert is a BOSH-trusted-certificate. You will need to update Ops Manager ca-certs (place the CA cert in /usr/local/share/ca-certificates and run update-ca-certificates, and restart tempest-web). You will need to add this certificate back in each time you do an upgrade of Ops Manager. In PCF 1.9+, Ops Manager will let you replace its own SSL cert and have that persist across upgrades.
- Update OpsManager File Storage Config to point at S3 blobstore using buckets (cc-buildpacks-<uniqueid>, cc-droplets-<uniqueid>, cc-packages-<uniqueid>, cc-resources-<uniqueid>)
- Click `Apply Changes` in Ops Manager
- Once changes are applied, re-run `goblob` to migrate any files which were created after the initial migration
- Validate apps can be restaged and pushed
- Update NFS Resource Config to 0 instances in Ops Manager to remove the NFS server. Also, you must set Cloud Controllers, Clock Global, and Cloud Controller Workers to 0 instances. This must be done to wipe the NFS mount from these Cloud Controller jobs. You can add these jobs back in later, after the NFS job is removed, but do note that this will incur downtime of Cloud Controller services (any `cf` commands) during this update.
- Click `Apply Changes` in Ops Manager.
- After the update is finished, you can add back the Cloud Controller, Clock Global, and Cloud Controller Worker instances in Ops Manager.
- Click `Apply Changes` in Ops Manager. After this is finished, availability of Cloud Controller services will resume.

## Developing

* Install [Docker](https://www.docker.com/products/docker)
* `docker pull minio/minio`

To run all of the tests in a Docker container:

`./testrunner`

To continually run the tests during development:

* `docker run -p 9000:9000 -e "MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE" -e "MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" minio/minio server /tmp`
* (in a separate terminal) `ginkgo watch -r`
