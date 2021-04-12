package main

import (
	"fmt"
	"obs"
)

var ak = "B8310C15008584EE0B90"
var sk = "R/wRjC+2V9yAAiG7zF8MaHRz/ncAAAF4AIWE70c0"
var endpoint = "http://hulian.com"

var obsClient, _ = obs.New(
	ak,
	sk,
	endpoint,
	obs.WithConnectTimeout(30),
	obs.WithSocketTimeout(60),
	obs.WithMaxConnections(100),
	obs.WithMaxRetryCount(3),
)

func main() {
	// input := &obs.GetObjectMetadataInput{}
	// input.Bucket = "test-bucket-0"
	// input.Key = "test-object-0"
	// // output, err := obsClient.HeadBucket("test-bucket-0")
	// output, err := obsClient.ListBuckets(nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(output)

	input := &obs.GetBucketMetadataInput{}
	input.Bucket = "test-bucket-0"
	output, err := obsClient.GetBucketMetadata(input)
	fmt.Println(output)
	if err == nil {
		fmt.Printf("RequestId:%s\n", output.RequestId)
		fmt.Printf("StorageClass:%s\n", output.StorageClass)
	} else {
		if obsError, ok := err.(obs.ObsError); ok {
			fmt.Printf("StatusCode:%d\n", obsError.StatusCode)
		} else {
			fmt.Println(err)
		}
	}

	obsClient.Close()
}
