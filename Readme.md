
# Emery

A **WORK IN PROGRESS** imageserver that is backed by S3, fronted by groupcache, and supports HMAC
signing and verification of URLs

## example

```bash
$ ./emery --listen=:8081 --groupcache-peers=:8082,:8083 --s3bucket myimage-test --key MYK3Yc00l --expiration 30m &
$ ./emery --listen=:8082 --groupcache-peers=:8081,:8083 --s3bucket myimage-test --key MYK3Yc00l --expiration 30m &
$ ./emery --listen=:8083 --groupcache-peers=:8081,:8082 --s3bucket myimage-test --key MYK3Yc00l --expiration 30m &
```

Open `http://localhost:8081/_sign/path/to/image.jpg?width=100`
it will redirect to `http://localhost:8081/longHMAChashHERE/path/to/image.jpg?expiration=UNIXmilliSTAMP&width=100`

Stats are available on `http://localhost:8081/stats`
