# Setting up Cloud Connectors for Cloudflare

This document walks you through setting up an API key for a Cloudflare Cloud Connector.

These instructions ask you to:

- Confirm that you have met the prerequisites below.
- Follow the API key creation and configuration steps below.

## Background for Cloudflare Cloud Connectors

Currently, the Cosmos Cloudflare Cloud Connectors scan for the following types of assets:

- Cloudflare DNS Zones
- Cloudflare DNS Records

The Cloudflare cloud connectors authenticate to your Cloudflare account via a provided API key. At this time, Cloudflare does not offer a federated authentication method. All API keys are encrypted both at rest and in transit.

## Cloudflare Prerequisites

As prerequisites to setting up a Cloudflare cloud connector, please ensure that your environment has the following:

1. A new or existing Cloudflare account
2. A user account with permissions to create and manage API keys

You will need to follow the steps described in this document for each Cloudflare account that you would like to create a connector for.

# Cloudflare Implementation Steps

On your Cloudflare account, you will need to perform the following steps:

1. Create an API Key
2. Assign `Read All Resources` permissions to the API key

To create an API key with the appropriate permissions, perform the following steps:

1. After signing in to Cloudflare, select the "Account API Keys" option in the left sidebar, under the "Manage Account" drop-down.
2. Select the "Create Token" button
3. On the following screen, select the "Use Template" button beside the "Read All Resources" option. This will allow you to automatically receive any future asset type updates that are made to the Cloudflare Cloud Connector. The minimum required permissions needed currently are:

- Zone - DNS - Read
- Zone - Zone - Read
- Account - Account Settings - Read

4. On the next screen, verify the permissions that this template is granting the Cosmos Cloud Connectors.
5. Select the "Continue to summary" button
6. Select the "Create Token" button

On the next screen, the API key will be available to copy. Please send this value, along with the account's ID to Bishop Fox via the Cosmos portal or your company's encrypted communication method of choice.
