# aws-authenticator
MFA authenticator for AWS Cli

Basic MFA authenticator with assume role feature.

### Usage:

    $ aws-authenticator

Then enter MFA token at the prompt.
  
Note: There should be a "getToken" profile in your `~/.aws/credentials` with at least two user policies:
  
* GetUser (for current user)
* Get user token with MFA

Like:

    {
        "Sid": "AllowUsersMFAForOwnCredentials",
        "Effect": "Allow",
        "Action": [
            "iam:GetSessionToken",
            "iam:GetUser"
        ],
        "Resource": [
            "arn:aws:iam::<AWS-ACCOUNT-ID>:mfa/${aws:username}"
        ]
    }

### Installation

- Install golang if it is not already on your machine. See: https://golang.org/doc/install

- Set GOPATH environment variable if it does not exist. Ex `export GOPATH=~/dev/go`

- Clone repository:
```
$ cd $GOPATH
$ git clone git@github.com:boynux/aws-authenticator.git
```

- Get dependencies
```
$ cd $GOPATH/aws-authenticator
$ go get
```

- Install the tool 
```$ go install```
