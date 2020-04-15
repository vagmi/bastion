# Bastion - A SSH Gateway tool for AWS deployments

Bastion is a SSH gateway tool built for AWS deployments. A common problem
that most SSH deployments face is how to manage SSH access across team
members and how to easily onboard or offboard developers and their access.
Bastion aims to solve the problem by using SSH certificates.

## Installation

TODO: instructions for downloading prebuilt packages for various platforms.

## Getting started

To get started with bastion you can initialize it using the following
command. Before running this ensure that you have `AWS_PROFILE` or
`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` set. Ensure that you also set
`AWS_REGION` or `AWS_DEFAULT_REGION` set as well.

### Initialize Bastion

This sets up bastion on an AWS account.

```
$ bastion init
```

This command does a few things.
* Setup a master key with AWS KMS that is specific to bastion Initialize a
* dynamodb table Generate a SSH key pair for SSH Certificate Generate a data
* key for encrypting and encrypt the SSH keypair and store it in dynamodb Setup
* bastion as a lambda handler Setup a IAM Policy document that allows execute
* permission on the lambda function

### Grant users access to bastion

Bastion allows specific IAM users to request signatures.

```
$ bastion grant iam-arn
```

This grants an IAM user access to the bastion lambda function

### Sign SSH user

As a developer, if you'd like to connect to a server then you need to first
obtain a certificate. Certificate has a validity period for a few minutes
that ensures that it cannot be abused.

```
$ $(bastion sign --ssh-add path_to_public_key) 
```

This command does a few things.

* Invoke the bastion lambda function and send the public key as an argument If
* the user is authorized to invoke the function, it returns a signed cert The
* `--ssh-add` param outputs it an ssh-add command that adds the certificate.
* Ensure that you have your private key added as well.

### Instance management

#### Configure SSH server

```
# on the server that you need to configure
$ sudo bastion configure
```

This edits the sshd configuration to accept the SSH certificate. Please note that
the IAM instance role should allow for calling the lambda function.

#### Create a jump host

To create a jump host you can use the following command.

``` 
$ bastion instance create --subnet-id=subnet-id --public-key path_to_pub.key
```

If a bastion instance is not present in the subnet, it creates a `t2.micro`
instance and configures the SSH instance using the `UserData` script to configure
ssh server to accept certificate authentication. It also sets the IAM instance role
with the policy that allows it call the bastion lambda function. It imports the public
key as a new keypair and associates it with the instance.

#### Connect to a jump host

To connect to a bastion jump host you can use the following command.

```
$ $(bastion connect --subnet-id subnet-id)
```

This outputs the appropriate ssh command and allows you to jump to the relevant host. You
can also configure the subnet id using the `BASTION_SUBNET_ID` environment variable. To jump from
a bastion host to another service you can use the following command.

```
$ $(bastion jump --subnet-id subnet-id) user@private-ip
```

This sets up agent forwarding and sets a proxy command to jump to specified host and run any ssh commands on it.