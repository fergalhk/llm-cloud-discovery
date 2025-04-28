# CloudControl API

This agent uses the AWS CloudControl API to retrieve resources.

### Usage

```bash
go run . -prompt '<prompt>'
```

For example, `go run . -prompt 'for each VPC, tell me how many EC2 instances are running in it'`.

### Examples:

```
> find all the details you can for all of the cloudfront distributions in my account, including the configuration of their origins. List this alongside the websites they are fronting.
The CloudFront distribution with ID `ET9WY5Z27O3V6` fronts the website "www.fergalkearns.com" and has a configuration that includes:

- Origins: A single origin managed by S3 bucket named `web-www.fergalkearns.com-content.s3.us-east-1.amazonaws.com`. The Origin Access Identity is not configured, indicating no specific permissions for access to this origin.

- Viewer Certificate: Uses ACM certificate with ARN `arn:aws:acm:us-east-1:616348090352:certificate/06d2894a-9510-4fec-9320-c6b0689c9122` for SSL.

- Default Cache Behavior:
  - Viewer Protocol Policy: `redirect-to-https`
  - Allowed Methods: `HEAD`, `GET`, `OPTIONS`
  - Smooth Streaming and GrpcConfig are disabled
  - TrustedSigners and TrustedKeyGroups aren't specified
  
- HTTP Version: `http2`
  
- Compressing is not enabled.

For this distribution, the website's content originates from an S3 bucket. There isn't any specific configuration for custom error responses or restrictions on Geo-restriction at the CloudFront level. 

If you have other CloudFront distributions or need details about the origins and their configurations, please let me know!
```