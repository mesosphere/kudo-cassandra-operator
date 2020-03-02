package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func DeleteFolderInS3(folderName string) error {

	// The session the S3 Uploader will use
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		// Specify profile to load for the session's config
		Profile: "327650738955_Mesosphere-PowerUser",

		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region: aws.String("us-west-2"),
		}},
	))

	bucket := "kudo-cassandra-backup-test"

	svc := s3.New(sess)
	res, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: &bucket,
		Prefix: &folderName,
	})
	if err != nil {
		return fmt.Errorf("failed to list bucket contents: %v\n", err)
	}

	fmt.Printf("Found %d items with prefix %s in bucket %s\n", len(res.Contents), folderName, bucket)

	if len(res.Contents) == 0 {
		return nil
	}

	objects := make([]*s3.ObjectIdentifier, len(res.Contents))
	for i, o := range res.Contents {
		o := o
		oi := s3.ObjectIdentifier{
			Key: o.Key,
		}

		objects[i] = &oi
	}

	delResult, err := svc.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: &bucket,
		Delete: &s3.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket entries: %v\n", err)
	}

	fmt.Printf("Deleted %d items with prefix %s in bucket %s\n", len(delResult.Deleted), folderName, bucket)

	return nil
}
