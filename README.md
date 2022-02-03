# bucketproxy
AWS S3 proxy

# Preconditions
1. Must have a valid AWS credentials file as per https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html
2. Edit a `config.json` as per your bucket settings

# Example config.json
```
{
    "bucket_name":"vas-bucket-1",
    "region": "eu-central-1",
    "server_port": "8080"
}
```
(The port must match the second port in the run docker image command below.)

# Build docker
` docker build . -t bucketproxy`

# Run docker image
`docker run -p 127.0.0.1:8080:8080 -v $(pwd)/config.json:/app/config.json -v ~/.aws/credentials:/root/.aws/credentials bucketproxy`
