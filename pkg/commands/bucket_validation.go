package commands

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

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
	reportEvery       int
	prefix            string
	retriesOnError    int
	bucketClient      objstore.Bucket
	objectNames       map[string]string
	objectContent     string
	logger            log.Logger
}

type retryingBucketClient struct {
	objstore.Bucket
	retries int
}

func (c *retryingBucketClient) withRetries(f func() error) error {
	var tries int
	for {
		err := f()
		if err == nil {
			return nil
		}
		tries++
		if tries >= c.retries {
			return err
		}
	}
}

func (c *retryingBucketClient) Upload(ctx context.Context, name string, r io.Reader) error {
	return c.withRetries(func() error { return c.Bucket.Upload(ctx, name, r) })
}

func (c *retryingBucketClient) Exists(ctx context.Context, name string) (bool, error) {
	var res bool
	var err error
	err = c.withRetries(func() error { res, err = c.Bucket.Exists(ctx, name); return err })
	return res, err
}

func (c *retryingBucketClient) Iter(ctx context.Context, dir string, f func(string) error, opts ...objstore.IterOption) error {
	return c.withRetries(func() error { return c.Bucket.Iter(ctx, dir, f, opts...) })
}

func (c *retryingBucketClient) Get(ctx context.Context, name string) (io.ReadCloser, error) {
	var res io.ReadCloser
	var err error
	err = c.withRetries(func() error { res, err = c.Bucket.Get(ctx, name); return err })
	return res, err
}

func (c *retryingBucketClient) Delete(ctx context.Context, name string) error {
	return c.withRetries(func() error { return c.Bucket.Delete(ctx, name) })
}

// Register is used to register the command to a parent command.
func (b *BucketValidationCommand) Register(app *kingpin.Application) {
	bvCmd := app.Command("bucket-validation", "Validate that object store bucket works correctly.")

	bvCmd.Command("validate", "Performs block upload/list/delete operations on a bucket to verify that it works correctly.").Action(b.validate)
	bvCmd.Flag("object-count", "Number of objects to create & delete").Default("2000").IntVar(&b.objectCount)
	bvCmd.Flag("report-every", "Every X operations a progress report gets printed").Default("100").IntVar(&b.reportEvery)
	bvCmd.Flag("test-runs", "Number of times we want to run the whole test").Default("1").IntVar(&b.testRuns)
	bvCmd.Flag("prefix", "path prefix to use for test objects in object store").Default("tenant").StringVar(&b.prefix)
	bvCmd.Flag("retries-on-error", "number of times we want to retry if object store returns error").Default("3").IntVar(&b.retriesOnError)
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
	b.logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	ctx := context.Background()

	bucketClient, err := bucket.NewClient(ctx, b.cfg, "testClient", b.logger, prometheus.DefaultRegisterer)
	if err != nil {
		return errors.Wrap(err, "failed to create the bucket client")
	}

	b.bucketClient = &retryingBucketClient{
		Bucket:  bucketClient,
		retries: b.retriesOnError,
	}

	for testRun := 0; testRun < b.testRuns; testRun++ {
		err = b.createTestObjects(ctx)
		if err != nil {
			return errors.Wrap(err, "error when uploading test data")
		}

		err = b.validateTestObjects(ctx)
		if err != nil {
			return errors.Wrap(err, "error when validating test data")
		}

		err = b.deleteTestObjects(ctx)
		if err != nil {
			return errors.Wrap(err, "error when deleting test data")
		}

		level.Info(b.logger).Log("testrun_successful", testRun+1)
	}

	return nil
}

func (b *BucketValidationCommand) report(phase string, completed int) {
	if completed == 0 || completed%b.reportEvery == 0 {
		level.Info(b.logger).Log("phase", phase, "completed", completed, "total", b.objectCount)
	}
}

func (b *BucketValidationCommand) setObjectNames() {
	b.objectNames = make(map[string]string, b.objectCount)
	for objectIdx := 0; objectIdx < b.objectCount; objectIdx++ {
		b.objectNames[fmt.Sprintf("%s/%05X/", b.prefix, objectIdx)] = "testfile"
	}
}

func (b *BucketValidationCommand) createTestObjects(ctx context.Context) error {
	iteration := 0
	for dirName, objectName := range b.objectNames {
		b.report("creating test objects", iteration)
		iteration++

		objectPath := dirName + objectName
		err := b.bucketClient.Upload(ctx, objectPath, strings.NewReader(b.objectContent))
		if err != nil {
			return errors.Wrapf(err, "failed to upload object (%s)", objectPath)
		}

		exists, err := b.bucketClient.Exists(ctx, objectPath)
		if err != nil {
			return errors.Wrapf(err, "failed to check if obj exists (%s)", objectPath)
		}
		if !exists {
			return errors.Errorf("Expected obj %s to exist, but it did not", objectPath)
		}
	}
	b.report("creating test objects", iteration)

	return nil
}

func (b *BucketValidationCommand) validateTestObjects(ctx context.Context) error {
	foundDirs := make(map[string]struct{}, b.objectCount)

	level.Info(b.logger).Log("phase", "listing test objects")

	err := b.bucketClient.Iter(ctx, b.prefix, func(dirName string) error {
		foundDirs[dirName] = struct{}{}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list objects")
	}

	iteration := 0
	for dirName, objectName := range b.objectNames {
		b.report("validating test objects", iteration)
		iteration++

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
	b.report("validating test objects", iteration)

	return nil
}

func (b *BucketValidationCommand) deleteTestObjects(ctx context.Context) error {
	iteration := 0
	for dirName, objectName := range b.objectNames {
		b.report("deleting test objects", iteration)
		iteration++

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
			return errors.Wrapf(err, "failed to delete obj (%s)", objectPath)
		}
		exists, err = b.bucketClient.Exists(ctx, objectPath)
		if err != nil {
			return errors.Wrapf(err, "failed to check if obj exists (%s)", objectPath)
		}
		if exists {
			return errors.Errorf("Expected obj %s to not exist, but it did", objectPath)
		}

		foundDeletedDir := false
		err = b.bucketClient.Iter(ctx, b.prefix, func(dirName string) error {
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
	b.report("deleting test objects", iteration)

	return nil
}
