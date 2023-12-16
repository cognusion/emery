package main

// A variant of his has been submitted as https://github.com/cognusion/imageserver/pull/28
// to be an official "source/s3". If that gets merged this will disappear.
import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cognusion/go-timings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/cognusion/imageserver"
)

// Server is a imageserver.Server implementation that gets the Image from an S3 URL.
//
// It parses the "source" param as URL, then do a GET request.
// It returns an error if the HTTP status code is not 200 (OK).
type Server struct {
	Session    *aws.Config
	BucketName string
}

// NewS3Server returns an s3.Server (imageserver.Server) capable of retrieving images from S3, or an error
func NewS3Server(awsRegion, awsAccessKey, awsSecretKey, awsS3Bucket string) (imageserver.Server, error) {
	awsSession, err := newAWSSession(awsRegion, awsAccessKey, awsSecretKey)
	if err != nil {
		return nil, err
	}

	return &Server{
		Session:    awsSession,
		BucketName: awsS3Bucket,
	}, nil
}

func newAWSSession(awsRegion, awsAccessKey, awsSecretKey string) (*aws.Config, error) {
	config := aws.NewConfig()

	// Region
	if awsRegion != "" {
		// CLI trumps
		config.Region = awsRegion
	} else if os.Getenv("AWS_DEFAULT_REGION") != "" {
		// Env is good, too
		config.Region = os.Getenv("AWS_DEFAULT_REGION")
	} else {
		return nil, fmt.Errorf("cannot find AWS region")
	}

	// Creds
	if awsAccessKey != "" && awsSecretKey != "" {
		// CLI trumps
		config.Credentials = credentials.NewStaticCredentialsProvider(
			awsAccessKey,
			awsSecretKey,
			"")
	} else if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		// Env is good, too
		config.Credentials = credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"")
	}
	return config, nil
}

// Get implements imageserver.Server.
func (srv *Server) Get(params imageserver.Params) (*imageserver.Image, error) {
	pline := fmt.Sprintf("Params: %+v", params)
	defer timings.Track(pline, time.Now(), TimingOut)

	DebugOut.Println(pline)
	bucketPath, err := params.GetString("source")
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(*srv.Session)

	// HEAD request
	hoo, err := client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: awsString(srv.BucketName),
		Key:    awsString(bucketPath),
	})
	if err != nil {
		return nil, err
	}

	// Determine the file type
	format := identifyFormat(*hoo.ContentType) // TODO this right
	if format == "" {
		return nil, fmt.Errorf("content-type of '%s' not a valid image type", *hoo.ContentType)
	}

	// pre-allocate in memory buffer, where headObject type is *s3.HeadObjectOutput
	buf := make([]byte, int(*hoo.ContentLength))
	// wrap with aws.WriteAtBuffer
	w := manager.NewWriteAtBuffer(buf)

	// GET it
	downloader := manager.NewDownloader(client)
	if _, err := downloader.Download(context.TODO(), w,
		&s3.GetObjectInput{
			Bucket: awsString(srv.BucketName),
			Key:    awsString(bucketPath),
		}); err != nil {
		return nil, err
	}

	// Return the Image
	return &imageserver.Image{
		Format: format,
		Data:   w.Bytes(),
	}, nil
}

// awsString mimics aws.String() as many AWS SDK functions
// demand *string and not string :shrug:
func awsString(v string) *string {
	return &v
}

// identifyFormat returns the right side of an "image/" content-type string,
// or empty
func identifyFormat(contentType string) string {
	if contentType == "" {
		return ""
	} else if !strings.HasPrefix(contentType, "image/") {
		return ""
	}

	return strings.TrimPrefix(contentType, "image/")
}
