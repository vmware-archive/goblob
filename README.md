# goblob

## To Run the migration
- Download latest release from releases page
- scp tool to NFS server
- make executable
``` chmod +x goblob-linux ```
- run to do a first pass
```./goblob-linux migrate --cf-identifier uniqueid --s3-endpoint XXXX --s3-region XXXX --s3-accesskey XXXX --s3-secretkey XXXX ```
- update OpsManager File Storage Config to point at S3 blobstore using buckets (cc-buildpacks-<uniqueid>, cc-droplets-<uniqueid>, cc-packages-<uniqueid>, cc-resources-<uniqueid>)
- Apply Changes in OpsManager
- Once Apply changes are complete re-run to to pickup delta from legacy NFS store
```./goblob-linux migrate --cf-identifier uniqueid --s3-endpoint XXXX --s3-region XXXX --s3-accesskey XXXX --s3-secretkey XXXX ```
- After validations (restage apps, etc):
  - Update NFS resource config to 0 in OpsManager to remove NFS server
  - Run apply changes in OpsManager

## To run S3 tests without ./testrunner need to startup a minio docker container

```
docker run -d -p 9000:9000 -e "MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE" -e "MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" minio/minio server /export
```



