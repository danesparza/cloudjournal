package cloudwatch

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/danesparza/cloudjournal/data"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/danesparza/cloudjournal/journal"
)

// Service encapsulates cloudwatch session and operations
type Service struct {
	DB           *data.Manager
	LogGroupName string
}

// WriteToLog writes the journal entries to the cloudwatch log stream for the unit
// Might want to handle errors similarly to
// https://github.com/devops-genuine/opentelemetry-collector-contrib/blob/e38594a148080bd0b102281b830505c4acb1b736/exporter/awsemfexporter/cwlog_client.go#L84-L118
func (service Service) WriteToLog(unit string, entries []journal.Entry) error {

	//	Get defaults:
	groupName := viper.GetString("cloudwatch.group")
	streamName := unit
	awsProfileName := viper.GetString("cloudwatch.profile")
	cloudwatchRegion := viper.GetString("cloudwatch.region")

	// Define the session - using SharedConfigState which forces file or env creds
	// See https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html for more information
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String(cloudwatchRegion)},
		Profile:           awsProfileName, /* Specify the profile to use in the credentials file */
	})
	if err != nil {
		log.WithFields(log.Fields{
			"unit":               unit,
			"cloudwatch.profile": awsProfileName,
		}).WithError(err).Error("unable to create AWS session for cloudwatch logs")
		return err
	}

	// Determine if we are authorized to access AWS with the credentials provided. This does not mean you have access to the
	// services required however.
	_, err = sts.New(sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.WithFields(log.Fields{
			"unit":               unit,
			"cloudwatch.profile": awsProfileName,
		}).WithError(err).Error("cannot validate aws credentials")
		return err
	}

	//	Create the cloudwatch logs service from the AWS session
	svc := cloudwatchlogs.New(sess)

	//	See if the log stream exists already
	resp, err := svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(groupName),
		LogStreamNamePrefix: aws.String(streamName),
	})

	//	If we got an error (maybe the stream doesn't exist yet)...
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			//	If it's because the log information doesn't exist ...
			case cloudwatchlogs.ErrCodeResourceNotFoundException:
				//	... Create the log group
				_, err = svc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
					LogGroupName: aws.String(groupName),
				})
				if err != nil {
					log.WithFields(log.Fields{
						"unit":               unit,
						"cloudwatch.profile": awsProfileName,
						"cloudwatch.group":   groupName,
					}).WithError(err).Error("can't create the log group")
					return err
				}

				//	.... Create the stream
				_, err = svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
					LogGroupName:  aws.String(groupName),
					LogStreamName: aws.String(streamName),
				})
				if err != nil {
					log.WithFields(log.Fields{
						"unit":               unit,
						"cloudwatch.profile": awsProfileName,
						"cloudwatch.group":   groupName,
					}).WithError(err).Error("can't create the log stream")
					return err
				}

				//	Try to get the response again:
				resp, err = svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
					LogGroupName:        aws.String(groupName),
					LogStreamNamePrefix: aws.String(streamName),
				})

				if err != nil {
					log.WithFields(log.Fields{
						"unit":               unit,
						"cloudwatch.profile": awsProfileName,
						"cloudwatch.group":   groupName,
					}).WithError(err).Error("this thing just doesn't want to work!")
					return err
				}
			default:

			}
		}
	}

	//	Get the next sequence token if it's available
	nextSequenceToken := ""
	if len(resp.LogStreams) > 0 {
		nextSequenceToken = *resp.LogStreams[0].UploadSequenceToken
	}

	// Create cloudwatch log events from our entries
	events := []*cloudwatchlogs.InputLogEvent{}
	for _, entry := range entries {

		//	Convert RealtimeTimestamp to an int64:
		timestamp, err := strconv.ParseInt(entry.RealtimeTimestamp, 10, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"unit":                    unit,
				"cloudwatch.profile":      awsProfileName,
				"cloudwatch.group":        groupName,
				"entry.RealtimeTimestamp": entry.RealtimeTimestamp,
			}).WithError(err).Error("problem converting timestamp to int64")
			continue
		}

		event := &cloudwatchlogs.InputLogEvent{
			Message:   aws.String(entry.Message),
			Timestamp: aws.Int64(int64(time.Nanosecond) * timestamp / int64(time.Millisecond)),
		}

		events = append(events, event)
	}

	//	Save all the events we gathered:
	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  aws.String(groupName),
		LogStreamName: aws.String(streamName),
		SequenceToken: &nextSequenceToken,
	}
	logResp, err := svc.PutLogEvents(params)
	if err != nil {
		log.WithFields(log.Fields{
			"unit":               unit,
			"cloudwatch.profile": awsProfileName,
			"cloudwatch.group":   groupName,
		}).WithError(err).Error("problem writing cloudwatch events")
		return err
	}

	//	Do we need to save this?
	nextSequenceToken = *logResp.NextSequenceToken

	return nil
}
