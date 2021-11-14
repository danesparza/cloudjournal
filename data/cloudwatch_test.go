package data_test

import (
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/sts"
)

// Rename to TestCloudwatch_WriteLogs_Successful to run as a test
func Cloudwatch_WriteLogs_Successful(t *testing.T) {

	//	This could come from environment
	group := "/app/cloudjournal"
	stream := "unittest"
	var nextSequenceToken string
	var m sync.Mutex

	// Define the session - using SharedConfigState which forces file or env creds
	// See https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html for more information
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String("us-east-1")},
		Profile:           "cloudjournal", /* Specify the profile to use in the credentials file */
	})
	if err != nil {
		t.Errorf("unable to create AWS session for cloudwatch logs: %v", err)
	}

	// Determine if we are authorized to access AWS with the credentials provided. This does not mean you have access to the
	// services required however.
	_, err = sts.New(sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		t.Errorf("cannot validate aws credentials: %v", err)
	}

	t.Logf("AWS session validated")

	//	Create the cloudwatch logs service from the AWS session
	svc := cloudwatchlogs.New(sess)

	//	See if the log stream exists already
	resp, err := svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(group),
		LogStreamNamePrefix: aws.String(stream),
	})

	//	If we got an error (maybe it didnt exist yet)...
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			//	If it's because the log information doesn't exist ...
			case cloudwatchlogs.ErrCodeResourceNotFoundException:
				//	... Create the log group
				_, err = svc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
					LogGroupName: aws.String(group),
				})
				if err != nil {
					t.Logf("Can't create the log group: %v", err)
				}

				//	.... Create the stream
				_, err = svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
					LogGroupName:  aws.String(group),
					LogStreamName: aws.String(stream),
				})
				if err != nil {
					t.Logf("Can't create the log stream: %v", err)
				}

				//	Try to get the response again:
				resp, err = svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
					LogGroupName:        aws.String(group),
					LogStreamNamePrefix: aws.String(stream),
				})

				if err != nil {
					t.Errorf("This thing just doesn't want to work! %v", err)
				}
			default:

			}
		}
	}

	if len(resp.LogStreams) > 0 {
		//	Get the next sequence token (WTF is this?)
		nextSequenceToken = *resp.LogStreams[0].UploadSequenceToken
		t.Logf("Next sequence token: %v", nextSequenceToken)
	}

	// Create a log message
	t.Logf("Logging a message...")
	event := &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(string("{\"message\": \"testing\", \"somethingelse\":\"overhere\"}")),
		Timestamp: aws.Int64(int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond)),
	}

	m.Lock()
	defer m.Unlock()

	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     []*cloudwatchlogs.InputLogEvent{event},
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		SequenceToken: &nextSequenceToken,
	}
	logResp, err := svc.PutLogEvents(params)
	if err != nil {
		t.Errorf("Error putting log events: %v", err)
	}

	nextSequenceToken = *logResp.NextSequenceToken
	t.Logf("Next sequence token: %v", nextSequenceToken)

}
