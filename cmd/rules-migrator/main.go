package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/cortexproject/cortex/pkg/chunk"
	"github.com/cortexproject/cortex/pkg/chunk/aws"
	"github.com/cortexproject/cortex/pkg/chunk/azure"
	"github.com/cortexproject/cortex/pkg/chunk/gcp"
	"github.com/cortexproject/cortex/pkg/chunk/openstack"
)

// ObjStoreConfig configures a rule store.
type ObjStoreConfig struct {
	Type  string                  `yaml:"type"`
	Azure azure.BlobStorageConfig `yaml:"azure"`
	GCS   gcp.GCSConfig           `yaml:"gcs"`
	S3    aws.S3Config            `yaml:"s3"`
	Swift openstack.SwiftConfig   `yaml:"swift"`
}

// RegisterFlags registers flags.
func (cfg *ObjStoreConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	cfg.Azure.RegisterFlagsWithPrefix(prefix, f)
	cfg.GCS.RegisterFlagsWithPrefix(prefix, f)
	cfg.S3.RegisterFlagsWithPrefix(prefix, f)
	cfg.Swift.RegisterFlagsWithPrefix(prefix, f)
	f.StringVar(&cfg.Type, prefix+"type", "gcs", "Method to use for backend rule storage (azure, gcs, s3)")
}

func main() {
	var (
		srcConfig = ObjStoreConfig{}
		dstConfig = ObjStoreConfig{}

		deleteSrc bool
	)
	flag.CommandLine.BoolVar(&deleteSrc, "delete-source", false, "If enabled, rule groups in the specified source store will be deleted upon migration.")
	srcConfig.RegisterFlagsWithPrefix("src.", flag.CommandLine)
	dstConfig.RegisterFlagsWithPrefix("dst.", flag.CommandLine)
	flag.Parse()

	srcClient, err := newClient(srcConfig)
	if err != nil {
		log.Fatalf("unable to initialize source bucket, %v", err)
	}

	dstClient, err := newClient(dstConfig)
	if err != nil {
		log.Fatalf("unable to initialize destination bucket, %v", err)
	}

	ctx := context.Background()

	log.Println("listing source rules")
	rgs, _, err := srcClient.List(ctx, "rules/", "")
	if err != nil {
		log.Fatalf("unable to list source rules, %v", err)
	}

	for _, rg := range rgs {
		newKey := generateRuleObjectKey(rg.Key)

		log.Printf("%s ==> %s\n", rg.Key, newKey)

		reader, err := srcClient.GetObject(ctx, rg.Key)
		if err != nil {
			log.Fatalf("unable to load object, %v", err)
		}

		data, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatalf("unable to read object, %v", err)
		}

		err = dstClient.PutObject(ctx, newKey, bytes.NewReader(data))
		if err != nil {
			log.Fatalf("unable to put object in destination bucket, %v", err)
		}

		if deleteSrc {
			err = srcClient.DeleteObject(ctx, rg.Key)
			if err != nil {
				log.Fatalf("unable to delete rule group from source bucket, %v", err)
			}
		}
	}
}

func newClient(cfg ObjStoreConfig) (chunk.ObjectClient, error) {
	switch cfg.Type {
	case "azure":
		return azure.NewBlobStorage(&cfg.Azure)
	case "gcs":
		return gcp.NewGCSObjectClient(context.Background(), cfg.GCS)
	case "s3":
		return aws.NewS3ObjectClient(cfg.S3)
	default:
		return nil, fmt.Errorf("Unrecognized rule storage mode %v, choose one of: configdb, gcs, s3, swift, azure", cfg.Type)
	}
}

func generateRuleObjectKey(key string) string {
	components := strings.Split(key, "/")
	if len(components) != 4 {
		panic(fmt.Sprintf("bad rule group found with '/' character, key='%s'; manual migration required", key))
	}
	return components[0] + "/" + components[1] + "/" + base64.URLEncoding.EncodeToString([]byte(components[2])) + "/" + base64.URLEncoding.EncodeToString([]byte(components[3]))
}
