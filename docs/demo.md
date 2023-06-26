# GreenGuardian Demo

## AWS Infrastructure

```shell
export AWS_PROFILE='GreenGuardianAdministrator-856591169022'

rm -rf crypto
mkdir -p crypto

aws iot create-policy --policy-name 'GreenGuardianGateway' --policy-document "$(cat <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "iot:Connect",
      "Resource": "arn:aws:iot:eu-north-1:856591169022:client/${iot:Connection.Thing.ThingName}"
    },
    {
      "Effect": "Allow",
      "Action": "iot:Publish",
      "Resource": "arn:aws:iot:eu-north-1:856591169022:topic//gateways/${iot:Connection.Thing.ThingName}/*"
    },
    {
      "Effect": "Allow",
      "Action": "iot:Subscribe",
      "Resource": "arn:aws:iot:eu-north-1:856591169022:topicfilter//gateways/${iot:Connection.Thing.ThingName}/*"
    },
    {
      "Effect": "Allow",
      "Action": "iot:Receive",
      "Resource": "arn:aws:iot:eu-north-1:856591169022:topic//gateways/${iot:Connection.Thing.ThingName}/*"
    }
  ]
}
EOF
)"

aws iot create-thing --thing-name 'DEVICE-Device_1'

aws iot create-keys-and-certificate --set-as-active --certificate-pem-outfile 'crypto/cert.pem' --private-key-outfile 'crypto/key.pem'

aws iot attach-policy --policy-name 'GreenGuardianGateway' --target 'arn:aws:iot:eu-north-1:856591169022:cert/feba75e6868feeed83897eb322b8b47ab656fc2a6c761b66bebbac60e312d2ae'
aws iot attach-thing-principal --thing-name 'DEVICE-Device_1' --principal 'arn:aws:iot:eu-north-1:856591169022:cert/feba75e6868feeed83897eb322b8b47ab656fc2a6c761b66bebbac60e312d2ae'

curl 'https://www.amazontrust.com/repository/AmazonRootCA1.pem' > 'crypto/ca.pem'

aws iot describe-endpoint --endpoint-type iot:Data-ATS
```

## Local Infrastructure

```shell
go run ./cmd/green-guardian-gateway/ --verbose
```

```shell
# We need root permissions because we're accessing the a USB device
# Be sure to plug in the IoT device beforeahand and consult `--help`
go build -o /tmp/green-guardian-hub ./cmd/green-guardian-hub/ && sudo /tmp/green-guardian-hub --verbose
```

## In-Person Demo

First, open [Cloudwatch](<https://eu-central-1.console.aws.amazon.com/cloudwatch/home?region=eu-central-1#logsV2:live-tail$3FlogGroupArns$3D~(~'arn*3aaws*3alogs*3aeu-central-1*3a097504859454*3alog-group*3aAWSIotLogsV2*3a*2a)>)

```shell
# Jumphost
export IP="49.13.2.195"
ssh -R 1337:localhost:1337 root@${IP}

# Gateway
ENDPOINT=ssl://ad218s2flbk57-ats.iot.eu-central-1.amazonaws.com:8883 THING_NAME=barn-tiger-ladle go run ./cmd/green-guardian-gateway/ --verbose

# Sensors (real)
export IP="49.13.2.195"
go build -o /tmp/green-guardian-hub ./cmd/green-guardian-hub/ && sudo /tmp/green-guardian-hub --verbose --raddr ${IP}:1337 --sprinklers='{}' --fans='{}' # --mock=50

# Actuators (mock)
export IP="49.13.2.195"
go build -o /tmp/green-guardian-hub ./cmd/green-guardian-hub/ && sudo /tmp/green-guardian-hub --verbose --raddr ${IP}:1337 --temperature-sensors='{}' --moisture-sensors='{}' --mock=50
```
