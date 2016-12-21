# goblob

## To execute tool
- Download latest release from releases page
- scp tool to NFS server
- chmod +x goblob-linux
- run
```
./goblob-linux migrate --cf-identifier uniqueid --s3-accesskey XXXX --s3-secretkey XXXX 
```

## To run S3 tests without ./testrunner need to startup a minio docker container

```
docker run -d -p 9000:9000 -e "MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE" -e "MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" minio/minio server /export
```



