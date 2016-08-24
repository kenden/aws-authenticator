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
(Including setting the GOPATH environment variable).

- Get the tool
````
$ go get github.com/boynux/aws-authenticator
````
- Install the tool  
It is already installed at $GOPATH/bin/aws-authenticator. 
To access it, you can either:
  * add the $GOPATH/bin to your PATH:  
  ```
  $ export PATH=$PATH:$GOPATH/bin
  ```
  (You can add this command to your .bshrc / .zshrc file to avoid typing it everytime you start a new shell).
  * copy the tool to a folder your PATH knows about. Ex:  
  ```
  $ cp $GOPATH/bin/aws-authenticator /usr/local/bin/
  ```
