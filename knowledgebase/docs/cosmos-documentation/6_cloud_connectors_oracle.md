# Setting up Cloud Connectors for Oracle

This document walks you through setting up a User with an API key for
an Oracle Cloud Connector.

These instructions ask you to:

- Confirm that you have met the prerequisites below.
- Follow the User and API key creation and configuration steps below.

## Background

Currently, the Oracle Cloud Connectors scan for the following types
of publicly facing assets:

- Oracle Buckets
- Oracle DNS Zones
- Oracle DNS Records
- Oracle Load Balancers
- Oracle Virtual Network Interface Card Attachments

## Oracle Prerequisites

As prerequisites to setting up an Oracle Cloud Connector, please
ensure that your environment has the following:

1. A new or existing Oracle Tenancy.
1. A user account with permission to create and manage Users and API keys.

## Oracle Implementation Steps

In your Oracle Tenancy Root Compartment, please perform the following steps.

Step 1: BishopFoxAudit Group Creation

1. Navigate to Identity & Security > Domains.
1. Choose an existing Domain. Note this domain for later.
1. Click on the "User management" Tab, Under the Groups section, press the "Create group" button.
1. BishopFox recommends the following values for clarity, but these can be whatever you prefer.
1. Name: `BishopFoxAudit`
1. Description: `Group for the BishopFoxAudit User`
1. Confirm "User can request access" is disabled.
1. No users will be assigned to this group at this time.
1. Add any additional tags if you prefer.

Press the "Create" button and continue to Step 2.

Step 2: BishopFoxAudit Policy Creation

1. Navigate to Identity & Security > Policies.
1. Press the "Create Policy" button.
1. BishopFox recommends the following values for clarity, but these can be whatever you prefer.
1. Name: `BishopFoxAudit`
1. Description: `Policy enabling the BishopFoxAudit Group/User`
1. Confirm the Compartment is your Tenancy Root Compartment.
1. Press the "Show manual editor" button. Using one of the following template options, replace
   `{domain}` with the Domain chosen in Step 1, and `{group}` with the Group created in Step 1.

Option 1, Recommended - To ensure that future updates to the BishopFox Oracle Cloud Connector
can continue to function without requiring policy updates.

```txt
Allow group '{domain}'/'{group}' to read all-resources in tenancy
```

Option 2, Minimum Required - The current minimum required policy statements necessary for the
BishopFox Oracle Cloud Connector functionality.

```txt
Allow group '{domain}'/'{group}' to inspect compartments in tenancy
Allow group '{domain}'/'{group}' to read buckets in tenancy
Allow group '{domain}'/'{group}' to read dns-zones in tenancy
Allow group '{domain}'/'{group}' to inspect vnic-attachments in tenancy
Allow group '{domain}'/'{group}' to inspect instances in tenancy
Allow group '{domain}'/'{group}' to inspect vnics in tenancy
Allow group '{domain}'/'{group}' to read load-balancers in tenancy
Allow group '{domain}'/'{group}' to read dns-records in tenancy
```

Press the "Create" button and continue to Step 3.

<!-- TODO: ^ minimum permissions ^ -->

Step 3: BishopFoxAudit User Creation

1. Navigate to Identity & Security > Domains.
1. Choose the same Domain from Step 1.
1. Click on the "User management" Tab, Under the Users section, press the "Create" button.
1. BishopFox recommends the following values for clarity, but these can be whatever you prefer.
1. Disable the "Use the email address as the username" setting.
1. First Name: `BishopFox`
1. Last Name: `Audit`
1. Username: `BishopFoxAudit`
1. Email: `bishopfox-audit@<your-organization>`
1. Select the checkbox for the Group created in Step 1.

Press the "Create" button and continue to Step 4.

Step 4: User API key Setup

1. Navigate to Identity & Security > Domains.
1. Choose the same Domain from Step 1.
1. Find the newly created user in the Users table, and click the link under the Username column.
1. Click on the "API keys" Tab
1. Press the "Add API key" button
1. Select the "Paste a public key" option.
1. BishopFox holds the Private Key paired with this Public Key.
   Paste this Public Key into the text box:

```txt
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA3VpKNkezocAM8u12RCAJ
QLUMt7iTcdPNEGTsX4EmurcyG7328y+bv4hmJKd0FUMG9Pl0FZXkZ2iSSJKiTzz3
9gnhKB0DSFlkn0eOln3ttS4tgBMXUc9SL4wXs4HzPfg8vHyh2WolJEcdznOdMLzp
53rtll7f4p9+lCRvwplDTkKOm9Bzk42IpW562ZEzie2RpR+8s/fSoG0uhOFH59nW
cga1QZNI9ysF4EymIxyxqYZsPoVEWPyHC4+w7w2Hld8SCZhEKQ/U7sD5CUmvJMZZ
a1fmqbgAeABs2DRZU7iNXnMq+7zW/Q5FHCGR+ZvC/7piapAETkvwNgpyU7ONWgfU
oBxOXS9e9niaD9HN0B5C28cJJwemUqbU/7T6EUsjlxZwQC4SE3lNk+f6HXFCHw2T
49qKBZ9h7qRCpkLRm31HnXPQCMQz7BRzkzqzIv/qfBOrv3SjloNPJntEgduVaSQE
HZpNqaGG+JZmpAiHF3kts4NXqtsMJLI6hn7u2fNMooUyfa1UnlwCfnfft/K01IHk
NIwPEKQ6RPuusjW66K9mXfDN79PiLwC476bxPA/TSOEGhUj3CA3V9MJmW2ebale7
OiNRYpciXA7lOQX/GepXt8+UWCjfLJTDE1eCNtMrBmFBvfgQujHI60sBRZLSjQXa
OZzFVB8QT33IHRoyI0M81XECAwEAAQ==
-----END PUBLIC KEY-----
```

Press the "Add" button. On the next screen there will be a Configuration File preview. The final
step is to send that Configuration File to your BishopFox SDM.
