# lambdaeip

## _Internet connectivity for your VPC-attached Lambda functions without a NAT Gateway_ 

### Background

I occasionally have serverless applications that need to be attached to VPCs 
*and* have access to the Internet. The standard solution for that is to deploy
NAT Gateways into each availability zone at a cost of $43/mo per zone. For that
money I could have invoked my function 215 million times. 

Today I learned there's a ~better~ different way, courtesy of [Chaz Schlarp][tweet].
I immediately went about generalising it so that I could set it and forget it in all
my personal environments.

### Deployment

You have two options. The first is deploying the following CloudFormation template:

```yaml
Transform: AWS::Serverless-2016-10-31

Resources:
  App:
    Type: AWS::Serverless::Application
    Properties:
      Location:
        ApplicationId: arn:aws:serverlessrepo:us-east-1:607481581596:applications/lambdaeip
        SemanticVersion: 0.1.0
      Parameters:
        VpcId: vpc-abc123
```

The second option is clicking [this link][console] to open the AWS web console,
fill in the VPC ID and click the _Deploy_ button. It should look like the
following screenshot:

![console screenshot](deploy.png)

### How it works

When a VPC-attached Lambda function is created, the Lambda service will
create a network interface for it. This issues an EventBridge event, which
triggers `lambdaeip` to execute and associate an Elastic IP address with that
ENI. It releases the EIP when the Lambda function's ENI is deleted (e.g. if the
function itself is deleted).

The way you identify whether a Lambda function should receive this special 
treatment is by associating a "sentinel" security group with it. Here's
a complete example of how to do that:

```yaml
Transform: AWS::Serverless-2016-10-31

Parameters:
  SubnetIds:
    Type: List<AWS::EC2::Subnet::Id>

  SentinelGroupId:
    Type: AWS::SSM::Parameter::Value<AWS::EC2::SecurityGroup::Id>
    Default: /lambdaeip/security-group-id

Resources:
  Function:
    Type: AWS::Serverless::Function
    Properties:
      Runtime: python3.8
      Handler: index.handler
      Timeout: 5
      VpcConfig:
        SecurityGroupIds: [!Ref SentinelGroupId, sg-whatever-else, sg-you-want]
        SubnetIds: !Ref SubnetIds
      InlineCode: |
        import urllib.request

        def handler(a, b):
          content = urllib.request.urlopen("https://www.cloudflare.com/cdn-cgi/trace").read()
          print(content)
```

### Caveats

Chaz says not to use this in production, but YOLO if you care that much about
saving tens of dollars a month, it's probably not _really_ a production env, right?!

[tweet]: https://twitter.com/schlarpc/status/1415393605330501632
[console]: https://console.aws.amazon.com/lambda/home?region=us-east-1#/create/app?applicationId=arn:aws:serverlessrepo:us-east-1:607481581596:applications/lambdaeip
