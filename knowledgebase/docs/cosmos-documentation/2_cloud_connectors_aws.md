# Using Cloud Connectors

Cloud connectors enable Cosmos to continuously and securely ingest asset data from cloud environments, such as AWS, Azure, Cloudflare, and GCP. By using cloud connectors, your organization gets near real-time visibility into your external attack surface. By integrating directly with your organization's cloud providers, Cosmos can discover ephemeral resources and unmanaged assets, as they happen, even in complex, multi-account environments.

Setting up a cloud connector ensures your asset inventory is complete, current, and enriched with metadata like account IDs, region information, and unique identifiers when available. This enables Cosmos to accurately attribute assets to your organization, group them by context, and prioritize them based on risk. With better visibility comes better detection. Cosmos can identify risky exposures as they appear.

For security teams, this means faster detection of changes to your cloud-based attack surface, fewer blind spots, and stronger coverage across your cloud footprint. With a cloud connector in place, Cosmos becomes your near-real-time lens into cloud asset sprawl, which empowers your organization to manage risk proactively and stay ahead of emerging threats.

# Benefits of Using Cloud Connectors in Cosmos

**Overview**

Accurately attributing cloud assets to an organization is a well-known challenge due to the shared nature of cloud infrastructure. Cosmos addresses this challenge through the use of Cloud Connectors.

## Why Attribution Is Difficult Without Cloud Connectors

Cloud environments often obscure direct asset ownership due to factors like:

- **Provider-based certificates** that are not customer-specific

- **Tenant isolation and privacy protections**

- **Dynamic and ephemeral infrastructure**, which can rotate frequently

As a result, linking these assets back to an organization without direct access can be challenging.

## How Cloud Connectors Help

Cloud Connectors offer a clear solution by:

- **Providing definitive attribution** of cloud-based assets to your organization

- **Bypassing visibility limitations** that can’t be resolved through external scanning alone

- **Syncing daily**, ensuring near real-time visibility into your cloud footprint

- **Enhancing the accuracy** of Cosmos attack surface inventory for cloud environments

This leads to a high-fidelity view of your externally facing attack surface and exposure risks and improves the reliability of detections, tracking, and prioritization of cloud-specific vulnerabilities.

**Summary**

For customers operating in AWS, Azure, or other major cloud environments, enabling a Cloud Connector is a key step in achieving full-spectrum visibility and attribution accuracy within Cosmos.

For assistance with setup or more information, reach out to your SDM.

# Setting up Cloud Connectors for Amazon Web Services (AWS)

This document walks through the process of setting up a role with proper permissions for the Bishop Fox AWS Cloud Connectors.

These instructions ask you to:

Confirm that you have met the prerequisites below.

Follow the role creation steps below.

## Background for AWS Cloud Connectors

Currently, the Cosmos AWS Cloud connectors scans for the following types of assets:

- Public facing DNS data (Route53) - these are public facing domains and subdomains.
- ENI - External Network Interfaces - this pulls all EC2, interface, and nat_gateway network_load_balancer IP data
- ELB and ELBv2 instances
- ElasticSearch instances
- S3 Buckets with public access

AWS Cloud Connectors can be set up either for individual accounts or for an entire AWS organization. AWS Cloud Connectors authenticate to your AWS environment by assuming a role that you create within your Account/Organization that is configured with a trust relationship that includes the Cosmos Production account.

When setting up an AWS connector for a single account, you will need to create the IAM role within that account. For full instructions, please see the "Single-account Connector Setup" section below.

When setting up an AWS connector for an organization, you will need to create such roles in every account within your organization. Bishop Fox will provide a CloudFormation stackset to simplify this process. For full instructions, please see the "Organization-level Connector Setup" section below.

In either case, you will need to provide the following information back to Bishop Fox:

- ARN of the created role (in the case of Organization-level connectors, this will be the role created in your management account)
- ID of the Organization or Account you will be targeting.

# Single-account Connector Setup

To create a single-account connector, you will need to create an IAM role within the account that can be assumed by the Cosmos Production account and that has the appropriate permissions necessary for the Cosmos Cloud Connector to function.

**Before creating this connector you will need an External ID created by Bishop Fox. Please reach out to your SDM for this. This will be in the UUIDv4 Format.**

## Account Prerequsities

To create an account connector the following are required:

- Permission to create IAM roles.

## Account AWS Console Instructions

After logging in to your AWS console, perform the following:

1. Navigate to the IAM Dashboard.
2. Select "Roles".
3. Press the "Create role" button.
4. Select "AWS Account" in the "Trusted Entity Type" section.
5. Select "Another AWS Account" in the "An AWS Account" selection.
6. Paste the account ID `058264235785` in the "Account ID" text field.
7. Select the "Require External ID" checkbox.
8. Paste the External ID provided by Bishop Fox in the "External ID" text field. This value will be unique for each connector and will take the form of a UUID. If you have not received an External ID from Bishop Fox, please reach out to your SDM. This value is not considered a secret.
9. Press the "Next" button.
10. On the next page, search for the `SecurityAudit` policy. This is a read-only policy designed to be provided to security testers and grants read access to most AWS resources. Bishop Fox recommends the use of the `SecurityAudit` policy to ensure that future updates to cloud connectors do not require you to adjust your role policy. However, if you prefer to use the minimal set of permissions necessary for cloud connector functionality, please see the section "Minimal Necessary IAM Permissions" below.
11. Select checkbox to the left of the `SecurityAudit` policy (or a policy that you create with the minimal permissions), then press the "Next" button.
12. Enter a role name and description. Bishop Fox recommends the name `bf-audit` or `bishopfox-auditor` for clarity, but this value can be whatever value you prefer.
13. Review the trusted entities policy. Ensure that it specifies the Principal as AWS account `058264235785` and that it includes a `Condition` block that tests against the appropriate External ID.
14. Review Permissions section. Ensure that it either is assigned the SecurityAudit policy or a policy that you created with the appropriate permissions.
15. Create any tags that you feel are appropriate for this role.
16. Press the "Create role" button to finalize this role's creation.

After the role has been created, send the role's ARN as well as the account ID to your Bishop Fox SDM.

## Account AWS CLI Instructions

Alternatively, the IAM role can be created via the AWS CLI. To create the role, first authenticate your CLI with your AWS account, then execute the following command. When executing this command, be sure to replace the string `{EXTERNAL_ID_HERE}` with the external ID provided by Bishop Fox. Additionally, you can re-name the role by changing the `bishopfox-auditor` string.

```
aws iam create-role --role-name bishopfox-auditor --assume-role-policy-document '{
  "Version": "2012-10-17",
  "statement": [
    {
      "effect": "allow",
      "principal": {
        "aws": "arn:aws:iam::058264235785:root"
      },
      "action": "sts:assumerole",
      "condition": {
        "stringequals": {
            "sts:externalid": "{EXTERNAL_ID_HERE}"
        }
      }
    }
  ]
}'
```

Next, attach the `SecurityAudit` policy to this role. This is a read-only policy designed to be provided to security testers and grants read access to most AWS resources.

If you changed the name of the role in the previous step, please ensure that you specify the same role name in this command.

`aws iam attach-role-policy --role-name bishofox-auditor --policy-arn arn:aws:iam::aws:policy/SecurityAudit`

Bishop Fox recommends the use of the `SecurityAudit` policy to ensure that future updates to cloud connectors do not require you to adjust your role policy. However, if you prefer to use the minimal set of permissions necessary for cloud connector functionality, please see the section "Minimal Necessary IAM Permissions" below.

If you choose to use a policy other than the `SecurityAudit` policy, replace the `SecurityAudit` policy's ARN with the ARN of the policy you have created.

After executing this command, your role should be fully set up. Please review the created role to ensure that it is set up with an Assume Role Policy that trusts the account `058264235785` and has the appropriate External ID, and that it has the appropriate policy attached.

After verifying the role, please provide the role's ARN and the account ID to your Bishop Fox SDM.

# Organization-level Connector Setup

To set up an Organization-level AWS cloud connector, you will need to set up IAM roles within each account of your organization that can be assumed by the Cosmos Cloud Connector. As this is a time-consuming exercise to perform manually, Bishop Fox will provide a CloudFormation StackSet that can create this role for you in all accounts automatically. Please note that this needs to be run as a **StackSet** and not a **Stack.**

**Before creating this connector you will need an External ID created by Bishop Fox. Please reach out to your SDM for this. This will be in the UUIDv4 Format.**

## Organization Prerequisites

To create an organization connector the following are required:

- Permission to create IAM roles.
- Access to an Organization's Delegated Administrator account.
- Ability to run StackSets in Organization's Delegated Administrator account.
- Ability to retrieve Organization id.

[Register a delegated administrator member account](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacksets-orgs-delegated-admin.html)

## Organization AWS Instructions

An example CloudFormation StackSet can be found below:

```
{
  "Description": "BishopFox security audit role created by CloudFormation StackSet.",
  "Parameters": {
    "roleName": {
      "Type": "String",
      "Description": "Role Name (*** Please don't change the default unless requested ***)",
      "Default": "bishopfox-auditor"
    },
    "externalId": {
      "Type": "String",
      "Description": "External Id"
    }
  },
  "Resources": {
    "BFAuditRole": {
      "Type": "AWS::IAM::Role",
      "Properties": {
        "RoleName": {
          "Ref": "roleName"
        },
        "Description": "BishopFox SecurityAudit role",
        "AssumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "AWS": "arn:aws:iam::058264235785:root"
              },
              "Action": "sts:AssumeRole",
              "Condition": {
                "StringEquals": {
                  "sts:ExternalId": {
                    "Ref": "externalId"
                  }
                }
              }
            }
          ]
        },
        "ManagedPolicyArns": [
          "arn:aws:iam::aws:policy/SecurityAudit"
        ]
      }
    }
  }
}
```

When applied from an Organization's Delegated Administrator account, this StackSet will create an IAM role in each account of the organization named `bishopfox-audit` by default that has the appropriate Assume Role policy and the built-in `SecurityAudit`policy.

**This StackSet must be applied from the Organization's Delegated Administrator account,** and it should be applied to your organization's Root. To find your Organization's Root ID from the AWS console, navigate to the "AWS Organizations" dashboard, select "AWS accounts" from the "Policy management" dropdown, and find the `Root` in the Organization panel. The Root ID should begin with `r-`.

To apply the StackSet from the AWS console:

1. Log in to your Organization's Delegated Administrator account.
2. Navigate to the CloudFormation dashboard, then select the "StackSets" option from the left navbar.
3. Press the "Create StackSet" button.
4. Configure your IAM Execution role if necessary, such as if you require elevated permissions.
5. Leave the "Template is Ready" option selected
6. Select the "Upload a template file" option in the "Specify Template" selection
7. Upload the Bishop Fox provided StackSet template under the "Upload a template file" file selection
8. Press the "Next" button
9. Name your StackSet instance and give it a description. Bishop Fox recommends choosing a name such as bishopfox-audit for clarity.
10. Enter the External ID provided by Bishop Fox in the externalId text field
11. At this time, you may re-name the role, though Bishop Fox recommends the use of bishopfox-audit for clarity.
12. Press the "Next" button.
13. Add tags and select the Execution configuration as necessary for your account. Check the acknowledgement that CloudFormation may create IAM resources. Press the Next button.
14. Select "Deploy new stacks" from the "Add Stacks to stack set" section. If this is not the first time that you have executed this StackSet, instead select "Import stacks to stack set".
15. Select "Deploy stacks in organizational units". In the "Organization numbers" text field, specify the Organization's Root ID found earlier.
16. Configure deployment options as necessary for your organization
17. Press the "Next" button. Review, then deploy the StackSet.

After the StackSet finishes applying, please send Bishop Fox the following:

- The ARN of the created role within your Management account. If the role ARN from a different account is sent, then the cloud connector may not have permissions to enumerate accounts and organizational units.
- The Organization ID of your organization. This value should begin with `o-`.

# Organizational Unit Connector Setup

To set up an Organizational Unit AWS cloud connector, you will need to set up IAM roles within each account of your organizational unit that can be assumed by the Cosmos Cloud Connector. As this is a time-consuming exercise to perform manually, Bishop Fox will provide a CloudFormation StackSet that can create this role for you in all accounts automatically. Please note that this needs to be run as a **StackSet** and not a **Stack.**

**Before creating this connector you will need an External ID created by Bishop Fox. Please reach out to your SDM for this. This will be in the UUIDv4 Format.**

## Organizational Unit Prerequisites

To create an organizational unit connector the following are required:

- Permission to create IAM roles.
- Access to an Organization's Delegated Administrator account.
- Ability to run StackSets in Organization's Delegated Administrator account.
- Ability to retrieve Organizational Unit id.

[Register a delegated administrator member account](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacksets-orgs-delegated-admin.html)

## Organizational Unit AWS Instructions

An example CloudFormation StackSet can be found below:

```
{
  "Description": "BishopFox security audit role created by CloudFormation StackSet.",
  "Parameters": {
    "roleName": {
      "Type": "String",
      "Description": "Role Name (*** Please don't change the default unless requested ***)",
      "Default": "bishopfox-auditor"
    },
    "externalId": {
      "Type": "String",
      "Description": "External Id"
    }
  },
  "Resources": {
    "BFAuditRole": {
      "Type": "AWS::IAM::Role",
      "Properties": {
        "RoleName": {
          "Ref": "roleName"
        },
        "Description": "BishopFox SecurityAudit role",
        "AssumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "AWS": "arn:aws:iam::058264235785:root"
              },
              "Action": "sts:AssumeRole",
              "Condition": {
                "StringEquals": {
                  "sts:ExternalId": {
                    "Ref": "externalId"
                  }
                }
              }
            }
          ]
        },
        "ManagedPolicyArns": [
          "arn:aws:iam::aws:policy/SecurityAudit"
        ]
      }
    }
  }
}
```

When applied from a Organization's Delegated Administrator account, this StackSet will create an IAM role in each account of the organization named `bishopfox-audit` by default that has the appropriate Assume Role policy and the built-in `SecurityAudit` policy.

**This StackSet must be applied from the Organization's Delegated Administrator account.** To find your Organizational Unit ID from the AWS console, navigate to the "AWS Organizations" dashboard, select "AWS accounts" from the "Policy management" dropdown, and find the `Organizational Unit` in the Organization panel. The Organizational Unit ID should begin with `ou-`.

To apply the StackSet from the AWS console:

1. Log in to your Organization's Delegated Administrator account.
2. Navigate to the CloudFormation dashboard, then select the "StackSets" option from the left navbar.
3. Press the "Create StackSet" button.
4. Configure your IAM Execution role if necessary, such as if you require elevated permissions.
5. Leave the "Template is Ready" option selected
6. Select the "Upload a template file" option in the "Specify Template" selection
7. Upload the Bishop Fox provided StackSet template under the "Upload a template file" file selection
8. Press the "Next" button
9. Name your StackSet instance and give it a description. Bishop Fox recommends choosing a name such as `bishopfox-audit` for clarity.
10. Enter the External ID provided by Bishop Fox in the `externalId` text field
11. At this time, you may re-name the role, though Bishop Fox recommends the use of `bishopfox-audit` for clarity.
12. Press the Next button.
13. Add tags and select the Execution configuration as necessary for your account. Check the acknowledgement that CloudFormation may create IAM resources. Press the Next button.
14. Select "Deploy new stacks" from the "Add Stacks to stack set" section. If this is not the first time that you have executed this StackSet, instead select "Import stacks to stack set".
15. Select "Deploy stacks in organizational units". In the "Organization numbers" text field, specify the Organizational Unit ID found earlier.
16. Configure deployment options as necessary for your organization
17. Press the Next button. Review, then deploy the StackSet.

After the StackSet finishes applying, please send Bishop Fox the following:

- The ARN of the created role within your Management account. If the role ARN from a different account is sent, then the cloud connector may not have permissions to enumerate accounts and organizational units.
- The Organizational Unit ID. It should begin with `ou-`.

## Minimal Necessary IAM Permissions

This document makes use of the built-in `SecurityAudit` policy. This is a read-only policy designed to be provided to security testers and grants read access to most AWS resources. Bishop Fox recommends the use of the `SecurityAudit` policy to ensure that future updates to cloud connectors do not require you to adjust your role policy.

In the case that your organization does not wish to provide this policy, this section details the minimal set of IAM permissions necessary for the Cloud Connector to function properly:

- `route53:ListHostedZones`
- `route53:ListResourceRecordSets`
- `ec2:DescribeNetworkInterfaces`
- `ec2:DescribeRegions`
- `elasticloadbalancing:DescribeLoadBalancers`
- `es:ListDomainNames`
- `es:DescribeElasticsearchDomain`
- `s3:ListAllMyBuckets`
- `s3:GetBucketPolicyStatus`

Additionally, for Organization-level connectors, the following permissions will be required on the role created within the Management account:

- `organizations:ListAccounts`
- `organizations:ListAccountsForParent`
- `organizations:ListOrganizationalUnitsForParent`
- `organizations:ListRoots`

These additional permissions are required to enumerate Accounts, Organizational Units, and Roots within the Organization.

To use these minimal permissions, create an IAM policy with these permissions and replace the `SecurityAudit` ARN with the ARN of the ARN of the created policy on all AWS CLI commands or AWS Console steps.
