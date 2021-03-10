package commands

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cortexproject/cortex/pkg/storage/bucket"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thanos-io/thanos/pkg/objstore"
	"gopkg.in/alecthomas/kingpin.v2"
)

// BucketValidationCommand is the kingpin command for bucket validation.
type BucketValidationCommand struct {
	cfg               bucket.Config
	s3SecretAccessKey string
	objectCount       int
	testRuns          int
	bucketClient      objstore.Bucket
	objectNames       map[string]string
	objectContent     string
	verbose           bool
}

// Register is used to register the command to a parent command.
func (b *BucketValidationCommand) Register(app *kingpin.Application) {
	bvCmd := app.Command("bucket-validation", "Validate that object store bucket works correctly.")

	bvCmd.Command("validate", "Performs block upload/list/delete operations on a bucket to verify that it works correctly.").Action(b.validate)
	bvCmd.Flag("object-count", "Number of objects to create & delete").Default("2000").IntVar(&b.objectCount)
	bvCmd.Flag("test-runs", "Number of times we want to run the whole test").Default("1").IntVar(&b.testRuns)
	bvCmd.Flag("verbose", "Log the bucket client output").BoolVar(&b.verbose)
	bvCmd.Flag("backend", "Backend type, can currently only be \"s3\"").Default("s3").StringVar(&b.cfg.Backend)
	bvCmd.Flag("s3.endpoint", "The S3 bucket endpoint. It could be an AWS S3 endpoint listed at https://docs.aws.amazon.com/general/latest/gr/s3.html or the address of an S3-compatible service in hostname:port format.").StringVar(&b.cfg.S3.Endpoint)
	bvCmd.Flag("s3.bucket-name", "S3 bucket name").StringVar(&b.cfg.S3.BucketName)
	bvCmd.Flag("s3.access-key-id", "S3 access key ID").StringVar(&b.cfg.S3.AccessKeyID)
	bvCmd.Flag("s3.secret-access-key", "S3 secret access key").StringVar(&b.s3SecretAccessKey)
	bvCmd.Flag("s3.insecure", "If enabled, use http:// for the S3 endpoint instead of https://. This could be useful in local dev/test environments while using an S3-compatible backend storage, like Minio.").BoolVar(&b.cfg.S3.Insecure)
	bvCmd.Flag("s3.signature-version", "The signature version to use for authenticating against S3. Supported values are: v2, v4").Default("v4").StringVar(&b.cfg.S3.SignatureVersion)
}

func (b *BucketValidationCommand) validate(k *kingpin.ParseContext) error {
	b.cfg.S3.SecretAccessKey.Set(b.s3SecretAccessKey)

	if b.cfg.Backend != "s3" {
		return errors.New("backend type must be \"s3\"")
	}

	err := b.cfg.Validate()
	if err != nil {
		return errors.Wrap(err, "config validation failed")
	}

	b.setObjectNames()
	b.objectContent = "testData"
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	ctx := context.Background()

	bucketClientLogger := log.NewNopLogger()
	if b.verbose {
		// Bucket client is quite verbose, only pass it the real logger if --verbose is set.
		bucketClientLogger = logger
	}
	b.bucketClient, err = bucket.NewClient(ctx, b.cfg, "testClient", bucketClientLogger, prometheus.DefaultRegisterer)
	if err != nil {
		return errors.Wrap(err, "failed to create the bucket client")
	}

	for testRun := 0; testRun < b.testRuns; testRun++ {
		err = b.createTestData(ctx)
		if err != nil {
			return errors.Wrap(err, "error when uploading test data")
		}

		err = b.validateTestData(ctx)
		if err != nil {
			return errors.Wrap(err, "error when validating test data")
		}

		err = b.deleteTestData(ctx)
		if err != nil {
			return errors.Wrap(err, "error when deleting test data")
		}

		level.Info(logger).Log("testrun_successful", testRun+1)
	}

	return nil
}

func (b *BucketValidationCommand) setObjectNames() {
	b.objectNames = make(map[string]string, b.objectCount)
	for objectIdx := 0; objectIdx < b.objectCount; objectIdx++ {
		b.objectNames[fmt.Sprintf("%05X/", objectIdx)] = "testfile"
	}
}

func (b *BucketValidationCommand) createTestData(ctx context.Context) error {
	for dirName, objectName := range b.objectNames {
		objectPath := dirName + objectName
		err := b.bucketClient.Upload(ctx, objectPath, bytes.NewReader([]byte(b.objectContent)))
		if err != nil {
			errors.Wrapf(err, "failed to upload object (%s)", objectPath)
		}

		exists, err := b.bucketClient.Exists(ctx, objectPath)
		if err != nil {
			errors.Wrapf(err, "failed to check if obj exists (%s)", objectPath)
		}
		if !exists {
			return errors.Errorf("Expected obj %s to exist, but it did not", objectPath)
		}
	}

	return nil
}

func (b *BucketValidationCommand) validateTestData(ctx context.Context) error {
	foundDirs := make(map[string]struct{}, b.objectCount)

	err := b.bucketClient.Iter(ctx, "", func(dirName string) error {
		foundDirs[dirName] = struct{}{}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list objects")
	}

	for dirName, objectName := range b.objectNames {
		if _, ok := foundDirs[dirName]; !ok {
			return fmt.Errorf("Expected directory did not exist (%s)", dirName)
		}

		objectPath := dirName + objectName
		reader, err := b.bucketClient.Get(ctx, objectPath)
		if err != nil {
			return errors.Wrapf(err, "failed to get object (%s)", objectPath)
		}

		content, err := ioutil.ReadAll(reader)
		if err != nil {
			return errors.Wrapf(err, "failed to read object (%s)", objectPath)
		}
		if string(content) != b.objectContent {
			return errors.Wrapf(err, "got invalid object content (%s)", objectPath)
		}
	}

	return nil
}

func (b *BucketValidationCommand) deleteTestData(ctx context.Context) error {
	for dirName, objectName := range b.objectNames {
		objectPath := dirName + objectName

		exists, err := b.bucketClient.Exists(ctx, objectPath)
		if err != nil {
			errors.Wrapf(err, "failed to check if obj exists (%s)", objectPath)
		}
		if !exists {
			return errors.Errorf("Expected obj %s to exist, but it did not", objectPath)
		}

		err = b.bucketClient.Delete(ctx, objectPath)
		if err != nil {
			errors.Wrapf(err, "failed to delete obj (%s)", objectPath)
		}
		exists, err = b.bucketClient.Exists(ctx, objectPath)
		if err != nil {
			errors.Wrapf(err, "failed to check if obj exists (%s)", objectPath)
		}
		if exists {
			return errors.Errorf("Expected obj %s to not exist, but it did", objectPath)
		}

		foundDeletedDir := false
		err = b.bucketClient.Iter(ctx, "", func(dirName string) error {
			if objectName == dirName {
				foundDeletedDir = true
			}
			return nil
		})
		if err != nil {
			return errors.Wrapf(err, "failed to list objects")
		}
		if foundDeletedDir {
			return errors.Errorf("List returned directory which is supposed to be deleted.")
		}
	}

	return nil
}
