# cloudjournal [![CircleCI](https://circleci.com/gh/danesparza/cloudjournal.svg?style=shield)](https://circleci.com/gh/danesparza/cloudjournal)
Journald to AWS cloudwatch log shipper.  Easy as pie 🥧

## Installing
### Prerequisites
Cloudjournal will use the AWS credentials on the machine its installed on.  The AWS docs suggest managing your credentials using a file called 'credentials' -- like `/root/.aws/credentials` ([more information on the AWS documentation site](https://docs.aws.amazon.com/sdkref/latest/guide/file-location.html)).  It should include the `cloudjournal` profile and it should look something like this: 

```
[cloudjournal]
aws_access_key_id = AWS_ACCESS_KEY_ID_HERE
aws_secret_access_key = aws_secret_access_key_here
```
The credentials for the `cloudjournal` profile should have the following AWS permissions:

```JSON
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams",
                "logs:PutLogEvents",
                "logs:GetLogEvents",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
```

### Installing the package
Get the latest .deb package for your architecture here: https://github.com/danesparza/cloudjournal/releases/latest  

Install it using 
`sudo dpkg -i cloudjournal-1.0.45_armhf.deb`

## Configuration
Configuration is done via /etc/cloudjournal/config.yaml.  Here is an example configuration file:

```yaml
server:
  port: 3010
  allowed-origins: "*"
datastore:
  system: /var/lib/cloudjournal/db/system.db
log:
  level: info
cloudwatch:
  region: "us-east-1"
  group: "/app/cloudjournal/{unit}"
  stream: "{machineid}"
monitor:  
  units: daydash, avahi-daemon
  interval: 10
```

`server` indicates where a runtime diagnostic interface is hosted.  It may be removed

`datastore` is where state information is stored for cloudjournal.  Defaults to ~/cloudjournal/db 

`log.level` can be debug, info, warn, error -- and it corresponds to your desired level of log verbosity.  Defaults to info

`cloudwatch.region` is the region you would like to log events to.  Your credentials should be for this region.  Defaults to us-east-1

`cloudwatch.group` is the log group name to use. Both groups and streams can have tokens in their name.  Defaults to /app/cloudjournal/{unit}

`cloudwatch.stream` is the log stream name to use.  Both groups and streams can have tokens in their name.  Defaults to {hostname}

`monitor.units` is a comma seperated list of units to monitor and sent to AWS Cloudwatch.  ***required***

`monitor.interval` is the number of minutes to wait between log batches.  Defaults to 1

### Tokens
There are several tokens you can use when naming `cloudwatch.group` or `cloudwatch.stream`:

`{hostname}` - This will be replaced with the contents of [/etc/hostname](https://man7.org/linux/man-pages/man1/hostname.1.html)

`{machineid}` - This will be replaced with the contents of [/etc/machine-id](https://www.man7.org/linux/man-pages/man5/machine-id.5.html)

`{unit}` - This will be replaced with the name of the current systemd unit being processed.

## Getting your app logs to cloudwatch
Getting your app log to cloudwatch is simple now: If your app is installed as a systemd unit, just output your logs to the console -- they'll automatically be added to journald under your systemd unit.  Then cloudwatch can take the logs for your journald unit and ship them to cloudwatch every few minutes.  

JSON logging is highly recommended because [AWS Cloudwatch can automatically parse JSON logs](https://aws.amazon.com/about-aws/whats-new/2015/01/20/amazon-cloudwatch-logs-json-log-format-support/) and [will provide structured log filters](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html) and searching.
