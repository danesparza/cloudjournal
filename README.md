# cloudjournal [![CircleCI](https://circleci.com/gh/danesparza/cloudjournal.svg?style=shield)](https://circleci.com/gh/danesparza/cloudjournal)
Journald to AWS cloudwatch log shipper


## Installing
Don't forget to add `/root/.aws/credentials` ([more information on the AWS documentation site](https://docs.aws.amazon.com/sdkref/latest/guide/file-location.html)).  It should look like this: 

```
[cloudjournal]
aws_access_key_id = AWS_ACCESS_KEY_ID_HERE
aws_secret_access_key = aws_secret_access_key_here
```

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
  stream: "{hostname}"
monitor:  
  units: daydash, avahi-daemon
  interval: 1
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

`{hostname}` - This will be replaced with the output from [/etc/hostname](https://man7.org/linux/man-pages/man1/hostname.1.html)

`{machineid}` - This will be replaced with the output from [/etc/machine-id](https://www.man7.org/linux/man-pages/man5/machine-id.5.html)

`{unit}` - This will be replaced with the name of the current systemd unit being monitored.
