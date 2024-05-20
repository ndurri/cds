package aws

import (
	"plugin"
)

const lambdaLib = "/opt/lib/lambda.so"
const awsLib = "/opt/lib/aws.so"

type LambdaClient struct {
	StartS3 func(func(string, string))
	StartSNS func(func(string))
	S3 func(func(string, string))
	SNS func(func(string))
	API func(func(map[string]string, string)int)
	APIv2 func(func(string, map[string]string, map[string]string, string)(int, map[string]string, string))
}

type S3Client struct {
	Get func(string, string)([]byte, error)
	Put func(string, string, []byte)error
}

type SNSClient struct {
	Put func(string, string)error
}

var Start = LambdaClient{
	S3: getSymbol(lambdaLib, "StartS3").(func(func(string, string))),
	SNS: getSymbol(lambdaLib, "StartSNS").(func(func(string))),
	API: getSymbol(lambdaLib, "StartAPI").(func(func(map[string]string, string)int)),
	APIv2: getSymbol(lambdaLib, "StartAPIv2").(func(func(string, map[string]string, map[string]string, string)(int, map[string]string, string))),
}
var Lambda = LambdaClient{
	StartS3: getSymbol(lambdaLib, "StartS3").(func(func(string, string))),
	StartSNS: getSymbol(lambdaLib, "StartSNS").(func(func(string))),
}
var Config = getSymbol(awsLib, "Config").(func()error)
var S3 = S3Client{
	Get: getSymbol(awsLib, "S3Get").(func(string, string)([]byte, error)),
	Put: getSymbol(awsLib, "S3Put").(func(string, string, []byte)error),
}
var SNS = SNSClient{
	Put: getSymbol(awsLib, "SNSPut").(func(string, string)error),
}

func getSymbol(libname, fname string) interface{} {
	lib, err := plugin.Open(libname)
	if err != nil {
		panic(err)
	}
	libfn, err := lib.Lookup(fname)
	if err != nil {
		panic(err)
	}
	return libfn
}
