# Setting up Cloud Connectors for Microsoft Azure

This document walks you through setting up federated credentials for an Azure managed identity for use by the Cosmos Azure Cloud Connector.

## Background for Azure Cloud Connectors

Currently, the Cosmos Azure Cloud Connectors scan for the following types of assets:

Azure DNS Zones
Azure DNS Records
The Azure Cloud Connectors authenticate to your Azure account via federated credentials to assume a Managed Identity with reader access to your Azure subscription.

The use of federated credentials allows the Cosmos platform to acquire short-lived Azure credentials without requiring a shared secret or other long-lived credentials.

## Azure Prerequisites

As prerequisites to setting up a Cosmos Azure cloud connector, please ensure that your environment has the following:

1. A new or existing Azure Tenant
2. At least one subscription with at least one Resource Group configured

Please select the guide that fits your situation:

- Customers with only a single (or very few) Azure subscriptions should use the [Single Subscription Guide](#single-subscription-connector-setup)
- Customers with many Azure subscriptions should use the [Multiple Subscription Guide](#multiple-subscription-connector-setup)
- Customers with an existing Azure connector who are migrating to our improved connector should use the [Migration Guide](#migrate-existing-connector)
- Customers with an existing operational Azure Cloud Connector who need to update the scopes of this App Registration/Service Principal should use the [Updating an Azure Connector](#updating-an-existing-enterprise-application)

# Multiple Subscription Connector Setup

On the Azure side, you will need to perform the following steps:

1. Create an App Registration
2. Add reader permissions to the App Registration as Service Principal Scopes.
3. Create Federated credentals for this App Registration

## Azure CLI

The Azure CLI will accomplish App Registration, reader permissions, and Service Principal setup simultaneously using the command below.

The example commands will create an App Registration named BishopFoxAudit, but it can be named however you desire.

The `--scopes` parameter may contain multiple scopes as a space-delimited list.

```
az ad sp create-for-rbac -n "BishopFoxAudit" --role reader --scopes /subscriptions/{subId}/resourceGroups/{resourceGroup}
```

Example with multiple scopes:

```
az ad sp create-for-rbac -n "BishopFoxAudit" --role reader --scopes \
  /subscriptions/{subId}/resourceGroups/{resourceGroup} \
  /subscriptions/{subIdTwo}/resourceGroups/{resourceGroupTwo} \
  /subscriptions/{subIdThree}/resourceGroups/{resourceGroupThree}
```

**NOTE:** Save the `appId` from the output above, it will be necessary for creating the Federated Credential. The`password` from the output above will be unused.

```
az ad app federated-credential create --id {appId-from-above} --parameters '{
  "name": "BishopFoxAuditFederatedCredentials",
  "issuer": "https://cognito-identity.amazonaws.com",
  "subject": "us-east-2:0d90d1af-0914-cd3e-0eb2-87560aebb179",
  "description": "Federated Credentials for the BishopFoxAudit App Registration.",
  "audiences": [
      "us-east-2:1a9f03f6-087c-42d3-b097-9a69af1ef906"
  ]
}'
```

This concludes setup from the Azure side of the multi-subscription Connector. The final step is to send the following details to your Bishop Fox SDM:

- Application (client) ID / `(appId)`
- Directory (tenant) ID
- Subscription ID to be used as a default

## Azure Portal

Instructions coming soon.

# Single Subscription Connector Setup

On the Azure side, you will need to perform the following steps:

1. Create a Managed Identity
2. Create Federated Credentials for the Cosmos Platform
3. Add reader permissions to the Managed Identity

## Creation of Managed Identity

Within the a resource group in the appropriate subscription, create a Managed Identity. Throughout this document, this role will be named `BishopFoxAudit`, but it can be named however you desire.

This can be done through the Azure web UI by performing the following steps:

1. Navigate to the `Managed Identity` page of the Azure portal
2. Press the `Create` button
3. Select the appropriate Subscription, Resource Group, and Region for your managed identity. The Managed Identity must be created in the Subscription in which you will be setting the connector up.
4. Name the managed identity something memorable. We suggest `BishopFoxAudit`.
5. If you wish to add tags, press Next, then add the desired tags and press the `Review + create` button. Otherwise, just press the `Review + create` button.
6. Review the summary, then press the `Create` button.

Alternatively, this can be done through the Azure CLI. An example follows:

```
az identity create --resource-group {resource_group} --name BishopFoxAudit
```

Replace the string `{resource_group}` with the name of the resource group in which you would like to create the managed identity. Additionally, as before, you can name the identity something other than `BishopFoxAudit` if desired.

## Create Federated Credentials

To allow the Cosmos platform to connect to managed identity, create a federated credential in the managed identity. When configuring this federated credential, you will need to set the following values to correspond with the Cosmos Platform's identity provider:

- Issuer URL: `https://cognito-identity.amazonaws.com`
- Subject identifier: `us-east-2:0d90d1af-0914-cd3e-0eb2-87560aebb179`
- Audience: `us-east-2:1a9f03f6-087c-42d3-b097-9a69af1ef906`

To set the federated credential up through the Azure web portal, perform the following steps:

1. Navigate to the Managed Identities page, then select the identity created in the previous step.
2. In the left side-bar, select the `Settings` drop-down, then select `Federated Credentials`.
3. Press the `+ Add Credential` button.
4. In the `Federated credential scenario` drop-down, select `Other`.
5. Fill in the `Issuer URL` and `Subject identifier` fields with the details above.
6. Provide a Name for the credential. This can be any value and is just for organizational purposes.
7. Under the `Audience` field, select `Enable (optional)`, then fill in the `Audience` field with the value from above.
8. Press the `Add` button.

Alternatively, this can be done through the Azure CLI. An example follows:

```
az identity federated-credential create --name cosmos-access \
--identity-name BishopFoxAudit --resource-group {resource_group} \
--issuer "https://cognito-identity.amazonaws.com" \
--subject "us-east-2:0d90d1af-0914-cd3e-0eb2-87560aebb179" \
--audiences "us-east-2:1a9f03f6-087c-42d3-b097-9a69af1ef906"
```

Be sure to replace the `{resource_group}` with the name of the resource group that the managed identity was created in, and if necessary replace `BishopFoxAudit` with the name of the managed identity created in the previous step. Additionally, this example will name the credential `cosmos-access`, but this can also be changed if desired.

## Grant Managed Identity Permissions

In order to function, the managed identity needs permissions to access your Azure subscription. We request that you provide the reader role so that the connector can ingest as much data as possible from the Azure subscription. If general reader permissions are scoped too broadly, you can create a custom role that requires only the needed permissions. If you take this approach, there will be additional maintenance of permissions as we add new asset types to our Azure Cloud Connector. Currently the minimum permissions required are:

- "Microsoft.Resources/subscriptions/resourceGroups/read"
- "Microsoft.Network/dnszones/read"
- "Microsoft.Network/dnszones/A/read"
- "Microsoft.Network/dnszones/AAAA/read"
- "Microsoft.Network/dnszones/CNAME/read"
- "Microsoft.Network/dnszones/NS/read"

To grant this permission from the web UI, perform the following steps:

1. Navigate to the managed identity created in the previous step
2. On the left sidebar, select `Azure role assignments`, then press the `+ Add role assignment` button.
3. In the `Scope` dropdown, select `Subscription`, then select the appropriate subscription from the Subscription dropdown. Alternatively, if you prefer to only grant access to a resource group, select `Resource Group` from the `Scope` dropdown and select the appropriate Subscription and Resource Group.
4. In the `Role` dropdown, select the `Reader` role. Alternatively you can use a custom role that includes the permissions listed above.
5. Press the `Save` button.

Alternatively, this can be done through the Azure CLI using either the `Reader` role or the custom role that has the required permissions. An example follows:

```
az role assignment create --role Reader \
--scope /subscriptions/{subscription_uuid} \
--assignee {managed_identity_uuid}
```

Alternatively, if you wish to only provide access to specific resource groups rather than the entire subscription, use the following form:

```
az role assignment create --role Reader \
--scope /subscriptions/{subscription_uuid}/resourceGroups/{resource_group_name} \
--assignee {managed_identity_uuid}
```

In this form, provide a `--scope` argument specifying the resource group path for each resource group you wish the cloud connectors to access.

Be sure to substitute `{managed_identity_uuid}`, `{subscription_uuid}`, and `{resource_group_name}` if appropriate with the corresponding value for your environment.

# Migrate Existing Connector

Migration steps consist of creating a Federated Credential under an existing App Registration set up from the previous Azure Cloud Connector implementation.

## Migrate using Azure CLI

Identify the `appId` of the existing Connector's App Registration, depending on how it was named:

`az ad app list --filter "startswith(displayName,'bishopfox')"`

Or locate it within the Azure Portal, the `appId` is labeled as "Application (client) ID".

```az ad app federated-credential create --id {appId-from-above} --parameters '{
  "name": "BishopFoxAuditFederatedCredentials",
  "issuer": "https://cognito-identity.amazonaws.com",
  "subject": "us-east-2:0d90d1af-0914-cd3e-0eb2-87560aebb179",
  "description": "Federated Credentials for the BishopFoxAudit App Registration.",
  "audiences": [
      "us-east-2:1a9f03f6-087c-42d3-b097-9a69af1ef906"
  ]
}'
```

This concludes setup from the Azure side of migrating an existing Connector. The final step is to send the following details to your Bishop Fox SDM:

- Application (client) ID / (`appId`)
- Directory (tenant) ID
- Subscription ID to be used as a default

## Migrate using Azure Portal

1. Locate the existing Connector's App Registration within the Portal.
2. Select `Certificates & secrets` in the sidebar.
3. Select the `Federated credentials (0)` tab and select `+ Add credential`
4. For `"Federated credential scenario"` choose `Other issuer` and enter the following:
5. Issuer: `https://cognito-identity.amazonaws.com`
6. Verify that the `Type` is selected as `Explicit subject identifier`
7. Value: `us-east-2:0d90d1af-0914-cd3e-0eb2-87560aebb179`
8. Name: `BishopFoxAuditFederatedCredentials`
9. Description: `Federated Credentials for the BishopFoxAudit App Registration`
10. Audience: `us-east-2:1a9f03f6-087c-42d3-b097-9a69af1ef906`
11. Click `E`dit (optional)` if this input is not interactive.

This concludes setup from the Azure side of migrating an existing Connector. The final step is to send the following details to your Bishop Fox SDM:

- Application (client) ID / (`appId`)
- Directory (tenant) ID
- Subscription ID to be used as a default

# Updating an Existing Enterprise Application

Over time you may need to update the scopes that your App Registration/Service Principal has access to.

```
az role assignment update --role-assignment {json}
```

To check the list of scopes that your App Registration/Service Principal has access to use the following command:

```
az role assignment list --assignee 00000000-0000-0000-0000-000000000000
```
