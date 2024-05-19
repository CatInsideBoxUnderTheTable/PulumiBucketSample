package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a new KMS key
		kmsKey, err := createKmsKey(ctx)
		if err != nil {
			return err
		}

		s3, err := createBucket(ctx, kmsKey)
		if err != nil {
			return err
		}

		enforcePublicAccessBlock(ctx, s3)
		if err != nil {
			return err
		}

		ctx.Export("bucketName", s3.ID())
		ctx.Export("kmsKeyArn", kmsKey.Arn)
		ctx.Export("kmsKeyAlias", pulumi.String("alias/myS3bucketKey"))
		return nil
	})
}

func createKmsKey(ctx *pulumi.Context) (*kms.Key, error) {
	kmsKey, err := kms.NewKey(ctx, "bucketKey", &kms.KeyArgs{
		Description: pulumi.String("KMS key for S3 bucket encryption"),
	})
	if err != nil {
		return nil, err
	}

	// Create a KMS Alias for the key
	_, err = kms.NewAlias(ctx, "myKeyAlias", &kms.AliasArgs{
		TargetKeyId: kmsKey.ID(),
		Name:        pulumi.String("alias/myS3bucketKey"),
	})
	if err != nil {
		return nil, err
	}
	return kmsKey, err
}

func enforcePublicAccessBlock(ctx *pulumi.Context, bucket *s3.Bucket) error {
	_, err := s3.NewBucketPublicAccessBlock(ctx, "myBucketPublicAccessBlock", &s3.BucketPublicAccessBlockArgs{
		Bucket:                bucket.ID(),
		BlockPublicAcls:       pulumi.Bool(true),
		BlockPublicPolicy:     pulumi.Bool(true),
		IgnorePublicAcls:      pulumi.Bool(true),
		RestrictPublicBuckets: pulumi.Bool(true),
	})
	if err != nil {
		return err
	}
	return nil
}

func createBucket(ctx *pulumi.Context, key *kms.Key) (*s3.Bucket, error) {
	versioning := s3.BucketVersioningArgs{
		Enabled: pulumi.Bool(true),
	}

	encryptionConfig := s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
		SseAlgorithm:   pulumi.String("aws:kms"),
		KmsMasterKeyId: key.KeyId,
	}

	encryption := s3.BucketServerSideEncryptionConfigurationArgs{
		Rule: &s3.BucketServerSideEncryptionConfigurationRuleArgs{
			ApplyServerSideEncryptionByDefault: encryptionConfig},
	}

	bucketArgs := s3.BucketArgs{
		ServerSideEncryptionConfiguration: &encryption,
		Versioning:                        &versioning,
	}

	return s3.NewBucket(ctx, "myBucket", &bucketArgs)
}
