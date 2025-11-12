# Setting up Cloud Connectors for Google Cloud Platform (GCP)

This document walks you through setting up workload identity federation for your Google Cloud Platform (GCP) so that Cosmos can interface with it and gather attack surface data for that GCP project.

These instructions ask you to:

- Confirm that you have met the prerequisites below.

- Follow the workload identity pool creation and configuration steps below.

## Background for GCP Cloud Connectors

Using workload identity federation, workloads that run on AWS EC2 can exchange their environment-specific credentials for short-lived Google Cloud Security Token Service tokens.

You can then use the same workload identity pool and provider for multiple workloads and across multiple Google Cloud projects.

Currently the GCP Cloud Connector scans for the following types of assets:

- Compute Instances
- Cloud Storage Buckets
- Forward Rules

## GCP Prerequisites

1. Have a new or existing Google Cloud project dedicated for hosting Workload Identity Federation (**billing must be enabled for the project** – Workload Identity Federation is free of charge - [Workforce Identity Federation | Google Cloud](https://cloud.google.com/workforce-identity-federation?hl=en#section-8) )

2. Enable the following APIs in this new or existing dedicated project:
   a. Identity and Access Management (IAM) API
   b. Cloud Resource Manager API
   c. IAM Service Account Credentials API
   d. Security Token Service API

[This link will enable all 4 APIs](https://accounts.google.com/v3/signin/identifier?continue=https%3A%2F%2Fconsole.cloud.google.com%2Fflows%2Fenableapi%3Fapiid%3Diam.googleapis.com%2Ccloudresourcemanager.googleapis.com%2Ciamcredentials.googleapis.com%2Csts.googleapis.com%26redirect%3Dhttps%3A%2F%2Fconsole.cloud.google.com%26_ga%3D2.132073423.1346845805.1686077570-381858120.1686077570&followup=https%3A%2F%2Fconsole.cloud.google.com%2Fflows%2Fenableapi%3Fapiid%3Diam.googleapis.com%2Ccloudresourcemanager.googleapis.com%2Ciamcredentials.googleapis.com%2Csts.googleapis.com%26redirect%3Dhttps%3A%2F%2Fconsole.cloud.google.com%26_ga%3D2.132073423.1346845805.1686077570-381858120.1686077570&ifkv=ASKXGp02gT698xWxWQ7ruROaSR_09-39Y4FGTB5Mo8-W9JD2v2uduxMTQz2Mz8WVRMgCJXtc4lut&osid=1&passive=1209600&service=cloudconsole&flowName=GlifWebSignIn&flowEntry=ServiceLogin&dsh=S-871480507%3A1701364404187626&theme=glif)

Enable the IAM, Resource Manager, Service Account Credentials, and Security Token Service APIs in this new or existing dedicated project

- [Enable the APIs](https://accounts.google.com/v3/signin/identifier?continue=https%3A%2F%2Fconsole.cloud.google.com%2Fflows%2Fenableapi%3Fapiid%3Diam.googleapis.com%2Ccloudresourcemanager.googleapis.com%2Ciamcredentials.googleapis.com%2Csts.googleapis.com%26redirect%3Dhttps%3A%2F%2Fconsole.cloud.google.com%26_ga%3D2.132073423.1346845805.1686077570-381858120.1686077570&followup=https%3A%2F%2Fconsole.cloud.google.com%2Fflows%2Fenableapi%3Fapiid%3Diam.googleapis.com%2Ccloudresourcemanager.googleapis.com%2Ciamcredentials.googleapis.com%2Csts.googleapis.com%26redirect%3Dhttps%3A%2F%2Fconsole.cloud.google.com%26_ga%3D2.132073423.1346845805.1686077570-381858120.1686077570&ifkv=ASKXGp02gT698xWxWQ7ruROaSR_09-39Y4FGTB5Mo8-W9JD2v2uduxMTQz2Mz8WVRMgCJXtc4lut&osid=1&passive=1209600&service=cloudconsole&flowName=GlifWebSignIn&flowEntry=ServiceLogin&dsh=S-871480507%3A1701364404187626&theme=glif)

Please confirm that you are in the correct project.

3. Have the following IAM roles on the dedicated project.
4. [roles/iam.workloadIdentityPoolAdmin](https://cloud.google.com/iam/docs/roles-permissions#iam.workloadIdentityPoolAdmin)
5. [roles/iam.serviceAccountAdmin](https://cloud.google.com/iam/docs/roles-permissions#iam.serviceAccountAdmin)

Alternatively, the IAM Owner (roles/owner) basic role also includes permissions to configure identity federation.

Have a service account already provisioned with at least Viewer role permissions to the projects that contains the assets you want scanned by Cosmos.

# GCP Implementation Steps

1. Create a new workload identity pool.

```bash
gcloud iam workload-identity-pools create POOL_ID \

    --location="global" \

    --description="DESCRIPTION" \

    --display-name="DISPLAY_NAME"
```

Replace the following values:

- POOL_ID: The unique ID for the pool, ex. \*\_bfidentitypool\_

- DESCRIPTION: The description of the pool (this description will display when granting access to pool identities)

- DISPLAY*NAME: The name of the pool, ex. *\_bishopfoxidentitypool\_

2. Add an AWS workload identity pool provider.

```bash
gcloud iam workload-identity-pools providers create-aws PROVIDER_ID \
--location="global" \
--workload-identity-pool="POOL_ID" \
--account-id="058264235785" \
--attribute-mapping="google.subject=assertion.arn,attribute.account=assertion.account,attribute.aws_role=assertion.arn.extract('assumed-role/cloud-connectors/cloud-connectors/aws-iam-role/BF_auditor/')"
```

Replace the following values:

- PROVIDER_ID: The unique ID for the provider, ex. \*\_bfawsprovider\_

- POOL_ID: The ID of the pool, ex. \*\_bfidentitypool\_

3. Within the same dedicated workload identity federation project, create the service account that will be connected to the workload identity pool via a pool identity.

```bash
gcloud iam service-accounts create SA_NAME \

    --description="DESCRIPTION" \

    --display-name="DISPLAY_NAME" \

    --project="PROJECT_ID"
```

4. Configure the external workload to impersonate the service account using the following steps. Obtain the project number of your current project, and execute the following command:

```
gcloud projects describe $(gcloud config get-value core/project) --format=value\(projectNumber\)
```

**Note**: Ensure you retrieve the project number of the dedicated project that corresponds to the project ID in the previous step.

- Grant the Workload Identity User role (roles/iam.workloadIdentityUser) to external identities that meet the required criteria (run each time for each AWS IAM Role):

  ```bash
  gcloud iam service-accounts add-iam-policy-binding SERVICE_ACCOUNT_EMAIL \

      --role=roles/iam.workloadIdentityUser \

      --member="principal://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL_ID/subject/arn:aws:sts::058264235785:assumed-role/BF_auditor/bf-cloud-connectors"
  ```

5. Create a credential configuration via the CLI with imdsv2 support.

```bash
  gcloud iam workload-identity-pools create-cred-config \

  projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL_ID/providers/PROVIDER_ID \

      --service-account=SERVICE_ACCOUNT_EMAIL \

      --aws \

      --enable-imdsv2 \

      --output-file=FILEPATH.json
```

Replace the following values:

- PROJECT_NUMBER: The project number of the project that contains the workload identity pool

- POOL_ID: The ID of the workload identity pool, ex. \_bfidentitypool\*

- PROVIDER_ID: The ID of the workload identity pool provider, ex. \_bfawsprovider\*

- SERVICE_ACCOUNT_EMAIL: The email address of the service account

- FILEPATH: The file to save configuration to

6. Assign the service account created in step #3 as a principal to the necessary GCP Organization, Folder(s) or Project(s). Assign the following roles to the principal:
   - Compute Viewer
   - Security Reviewer

If you prefer to use a minimally scoped role, please follow the Custom Role Creation Guide at the end of this guide.

If you want scanning at the organization level, you will also need to grant the following roles at the organization level:

- Security Viewer
- Folder Viewer (This is needed to provide the `resourcemanager.projects.get` permission which can also be added on its own.)

7. Finally, provide the following to Bishop Fox

- Workload Identity Federation credentials file that was created during step #5

- GCP Organization ID – run the following command:

```
  gcloud organizations list
```

## Custom Role Creation

This document walks you through setting up a custom role for your Google Cloud Platform (GCP). This is for customers who prefer a minimally scoped role to a GCP standard role. Using this guide may involve more maintenance as features are added to the GCP Cloud Connector.

### Implementation Steps for Org Scan

1. Create a custom role in your Organization

```bash
gcloud iam roles create ROLE_ID --organization ORG_ID
--description="DESCRIPTION" \
--permissions=PERMISSIONS
```

Replace the following values:

ROLE_ID is the name you want for the custom role ORG_ID is the id number for your GCP Org DESCRIPTION can be anything PERMISSIONS should be "resourcemanager.projects.list,resourcemanager.projects.get,storage.buckets.list,serviceusage.services.list,compute.instances.list,compute.forwardingRules.list,storage.buckets.getIamPolicy,resourcemanager.folders.list"

1. Attach the role to the existing Service Account

```bash
gcloud organizations add-iam-policy-binding ORG_ID
--member=serviceAccount:SERVICE_ACCOUNT_EMAIL \
--role=organizations/ORG_ID/roles/ROLE_ID
```

Replace the following values:

ORG_ID is the id number for your GCP Org SERVICE_ACCOUNT is the email address of the service account ROLE_ID is the name of the custom role

Check the GCP IAM tab to see if the service account has the new role attached

### Implementation Steps for Project Scan

1. Createa a custom role in your Project

```bash
gcloud iam roles create ROLE_ID --project PROJECT_ID \
--description="DESCRIPTION" \
--permissions=PERMISSIONS
```

Replace the following values:

ROLE_ID is the name you want for the custom role ORG_ID is the id number for your GCP Org DESCRIPTION can be anything PERMISSIONS should be "storage.buckets.list,compute.instances.list,compute.forwardingRules.list,storage.buckets.getIamPolicy,resourcemanager.projects.get,serviceusage.services.list"

-Note this list of permissions will need to be updated as we add more asset types to our scans

1. Attach the role to the existing Service Account

```bash
gcloud organizations add-iam-policy-binding PROJECT_ID
--member=serviceAccount:SERVICE_ACCOUNT_EMAIL \
--role=projects/PROJECT_ID/roles/ROLE_ID
```

Replace the following values:

PROJECT_ID is the alphanumeric name for the project SERVICE_ACCOUNT is the email address of the service account ROLE_ID is the name of the custom role

Check the GCP IAM tab to see if the service account has the new role attached
