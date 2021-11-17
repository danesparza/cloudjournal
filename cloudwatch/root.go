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

// GetAWSSession gets an AWS session to use with an operation
func (service Service) GetAWSSession() (*session.Session, error) {

	//	Get the configuration information for the AWS profile and region
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
			"cloudwatch.profile": awsProfileName,
			"cloudwatch.region":  cloudwatchRegion,
		}).WithError(err).Error("unable to create AWS session for cloudwatch logs")
		return nil, err
	}

	// Determine if we are authorized to access AWS with the credentials provided. This does not mean you have access to the
	// services required however.
	_, err = sts.New(sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.WithFields(log.Fields{
			"cloudwatch.profile": awsProfileName,
			"cloudwatch.region":  cloudwatchRegion,
		}).WithError(err).Error("cannot validate aws credentials")
		return nil, err
	}

	return sess, nil
}

// CreateLogGroup creates a cloudwatch log group
func (service Service) CreateLogGroup() error {
	//	Get an AWS session
	sess, err := service.GetAWSSession()
	if err != nil {
		log.WithFields(log.Fields{
			"cloudwatch.group": service.LogGroupName,
		}).WithError(err).Error("unable to create AWS session in order to create a log group")
		return err
	}

	//	Create the cloudwatch logs service from the AWS session
	svc := cloudwatchlogs.New(sess)

	//	.... Create the stream
	_, err = svc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(service.LogGroupName),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"cloudwatch.group": service.LogGroupName,
		}).WithError(err).Error("can't create the log group")
		return err
	}

	return nil
}

// CreateLogStream creates a cloudwatch log stream
func (service Service) CreateLogStream(streamName string) error {

	//	Get an AWS session
	sess, err := service.GetAWSSession()
	if err != nil {
		log.WithFields(log.Fields{
			"streamName": streamName,
		}).WithError(err).Error("unable to create AWS session in order to create a log stream")
		return err
	}

	//	Create the cloudwatch logs service from the AWS session
	svc := cloudwatchlogs.New(sess)

	//	.... Create the stream
	_, err = svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(service.LogGroupName),
		LogStreamName: aws.String(streamName),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"streamName":       streamName,
			"cloudwatch.group": service.LogGroupName,
		}).WithError(err).Error("can't create the log stream")
		return err
	}

	return nil
}

// WriteToLog writes the journal entries to the cloudwatch log stream for the unit
// Might want to handle errors similarly to
// https://github.com/devops-genuine/opentelemetry-collector-contrib/blob/e38594a148080bd0b102281b830505c4acb1b736/exporter/awsemfexporter/cwlog_client.go#L84-L118
func (service Service) WriteToLog(unit string, entries []journal.Entry) error {

	log.WithFields(log.Fields{
		"unit":       unit,
		"founditems": len(entries),
	}).Debug("requested write of items to cloudwatch logs")

	//	Get defaults:
	groupName := service.LogGroupName
	streamName := unit

	//	Get an AWS session
	sess, err := service.GetAWSSession()
	if err != nil {
		log.WithFields(log.Fields{
			"unit": unit,
		}).WithError(err).Error("unable to create AWS session in order to write a log message")
		return err
	}

	//	Create the cloudwatch logs service from the AWS session
	svc := cloudwatchlogs.New(sess)

	//	See if the log stream exists already
	resp, err := svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(groupName),
		LogStreamNamePrefix: aws.String(streamName),
	})

	//	If we got an error (or if we appear to have no log streams)...
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			//	If it's because the log information doesn't exist ...
			case cloudwatchlogs.ErrCodeResourceNotFoundException:
				log.WithFields(log.Fields{
					"aerr.code":    aerr.Code(),
					"aerr.message": aerr.Message(),
				}).WithError(err).Debug("describe log streams says the resource doesn't exist.  Creating the log group and log stream")

				//	Create the log group
				err = service.CreateLogGroup()
				if err != nil {
					log.WithError(err).Error("problem creating log group")
				}

				//	Create the log stream
				err = service.CreateLogStream(unit)
				if err != nil {
					log.WithError(err).Error("problem creating log stream")
				}

			default:
				log.WithFields(log.Fields{
					"aerr.code":    aerr.Code(),
					"aerr.message": aerr.Message(),
				}).WithError(err).Error("some other aws error is happening")
			}
		}
	}

	//	If we don't have log streams...
	if len(resp.LogStreams) < 1 {
		log.WithFields(log.Fields{
			"unit":             unit,
			"cloudwatch.group": groupName,
		}).Debug("we appear to have no log stream.  Attempting to create")

		//	Create the log stream
		err = service.CreateLogStream(unit)
		if err != nil {
			log.WithError(err).Error("problem creating log stream")
		}
	}

	//	Get the next sequence token if it's available
	nextSequenceToken := ""
	if len(resp.LogStreams) > 0 {
		if resp.LogStreams[0].UploadSequenceToken != nil {
			nextSequenceToken = *resp.LogStreams[0].UploadSequenceToken
			log.WithFields(log.Fields{
				"unit":              unit,
				"cloudwatch.group":  groupName,
				"nextSequenceToken": nextSequenceToken,
				"logStreamName":     *resp.LogStreams[0].LogStreamName,
			}).Debug("found next sequence token for logstream")
		} else {
			log.WithFields(log.Fields{
				"unit":             unit,
				"cloudwatch.group": groupName,
				"logStreamName":    *resp.LogStreams[0].LogStreamName,
			}).Debug("we found a logstream, but don't have a sequence token")
		}
	}

	// Create cloudwatch log events from our entries
	events := []*cloudwatchlogs.InputLogEvent{}
	for _, entry := range entries {

		//	Convert RealtimeTimestamp to an int64:
		timestamp, err := strconv.ParseInt(entry.RealtimeTimestamp, 10, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"unit":                    unit,
				"cloudwatch.group":        groupName,
				"entry.RealtimeTimestamp": entry.RealtimeTimestamp,
			}).WithError(err).Error("problem converting timestamp to int64")
			continue
		}

		//	Format the timestamp
		formattedTimestamp := int64(time.Nanosecond) * timestamp / int64(time.Millisecond)

		log.WithFields(log.Fields{
			"unit":    unit,
			"stream":  streamName,
			"tstamp":  formattedTimestamp,
			"message": entry.Message,
			"group":   groupName,
		}).Debug("adding log event")

		//	Add the event
		event := &cloudwatchlogs.InputLogEvent{
			Message:   aws.String(entry.Message),
			Timestamp: aws.Int64(formattedTimestamp),
		}

		events = append(events, event)
	}

	//	Format our log request
	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  aws.String(groupName),
		LogStreamName: aws.String(streamName),
	}

	//	If we have a sequence token, use it:
	if len(nextSequenceToken) > 0 {
		params.SequenceToken = &nextSequenceToken
	}

	//	Log our events
	log.WithFields(log.Fields{
		"unit":              unit,
		"streamName":        streamName,
		"nextSequenceToken": nextSequenceToken,
		"cloudwatch.group":  groupName,
		"eventCount":        len(params.LogEvents),
	}).Debug("writing to cloudwatch logs...")

	_, err = svc.PutLogEvents(params)
	if err != nil {
		log.WithFields(log.Fields{
			"unit":              unit,
			"streamName":        streamName,
			"nextSequenceToken": nextSequenceToken,
			"cloudwatch.group":  groupName,
		}).WithError(err).Error("problem writing to cloudwatch logs")
		return err
	}

	return nil
}
