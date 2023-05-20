# AWS Docs

```shell
export AWS_PROFILE='GreenGuardianAdministrator-856591169022'

rm -rf crypto
mkdir -p crypto

aws iot create-policy --policy-name 'GreenGuardianGatewayConnect' --policy-document "$(cat <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "iot:Connect",
      "Resource": "*"
    }
  ]
}
EOF
)"

aws iot create-policy --policy-name 'GreenGuardianGatewayPublish' --policy-document "$(cat <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "iot:Publish",
      "Resource": "arn:aws:iot:eu-north-1:856591169022:topic/gateways/${iot:Connection.Thing.ThingName}/*"
    }
  ]
}
EOF
)"

aws iot create-thing --thing-name 'GreenGuardianGateway1'

aws iot create-keys-and-certificate --set-as-active --certificate-pem-outfile 'crypto/aws.crt' --private-key-outfile 'crypto/aws.key'

aws iot attach-policy --policy-name 'GreenGuardianGatewayConnect' --target 'arn:aws:iot:eu-north-1:856591169022:cert/1bea93461bad943ca994d8d7b44a67e973239b52e7048ec0f3a8b59250999e16'
aws iot attach-policy --policy-name 'GreenGuardianGatewayPublish' --target 'arn:aws:iot:eu-north-1:856591169022:cert/1bea93461bad943ca994d8d7b44a67e973239b52e7048ec0f3a8b59250999e16'
aws iot attach-thing-principal --thing-name 'GreenGuardianGateway1' --principal 'arn:aws:iot:eu-north-1:856591169022:cert/1bea93461bad943ca994d8d7b44a67e973239b52e7048ec0f3a8b59250999e16'

curl 'https://www.amazontrust.com/repository/AmazonRootCA1.pem' > 'crypto/aws-ca.pem'

aws iot describe-endpoint --endpoint-type iot:Data-ATS
```
